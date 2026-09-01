package terminal

import (
	"io"
	"strings"
	"unicode"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// RenderPlain renders a durable document without styles or terminal control.
// Plain Interactive and Automation modes intentionally share this renderer.
func RenderPlain(document PresentationDocument) string {
	var output strings.Builder
	for index, block := range document.Blocks {
		if index > 0 && output.Len() > 0 && !strings.HasSuffix(output.String(), "\n") {
			output.WriteByte('\n')
		}
		text := block.Text
		if block.Sensitive {
			text = "[redacted]"
		}
		output.WriteString(stripTerminalControl(text))
	}
	if output.Len() > 0 && !strings.HasSuffix(output.String(), "\n") {
		output.WriteByte('\n')
	}
	return output.String()
}

// WritePlain writes a durable Command Result to its stdout destination.
func WritePlain(output io.Writer, document PresentationDocument) error {
	return writeComplete(output, RenderPlain(document))
}

// RichOptions controls terminal-owned Rich Interactive presentation behavior.
type RichOptions struct {
	Width int
	Color bool
}

// WriteRich writes a durable Rich Interactive Command Result to stdout.
func WriteRich(output io.Writer, document PresentationDocument, options RichOptions) error {
	return writeComplete(output, renderRich(document, options))
}

func writeComplete(output io.Writer, value string) error {
	written, err := io.WriteString(output, value)
	if err != nil {
		return err
	}
	if written != len(value) {
		return io.ErrShortWrite
	}
	return nil
}

func renderRich(document PresentationDocument, options RichOptions) string {
	var output strings.Builder
	styles := richStyles(options.Color)
	for index, block := range document.Blocks {
		if index > 0 && output.Len() > 0 && !strings.HasSuffix(output.String(), "\n") {
			output.WriteByte('\n')
		}
		text := block.Text
		if block.Sensitive {
			text = "[redacted]"
		}
		text = wrapText(stripTerminalControl(text), options.Width)
		output.WriteString(styles[block.Role].Render(text))
	}
	if len(document.Blocks) > 0 && !strings.HasSuffix(output.String(), "\n") {
		output.WriteByte('\n')
	}
	return output.String()
}

func richStyles(color bool) map[VisualRole]lipgloss.Style {
	plain := lipgloss.NewStyle()
	styles := map[VisualRole]lipgloss.Style{
		VisualRolePlain:   plain,
		VisualRoleTitle:   plain,
		VisualRoleActive:  plain,
		VisualRoleSuccess: plain,
		VisualRoleWarning: plain,
		VisualRoleError:   plain,
		VisualRoleMuted:   plain,
	}
	if !color {
		return styles
	}

	styles[VisualRoleTitle] = plain.Foreground(lipgloss.Color("6")).Bold(true)
	styles[VisualRoleActive] = plain.Foreground(lipgloss.Color("6")).Bold(true)
	styles[VisualRoleSuccess] = plain.Foreground(lipgloss.Color("10"))
	styles[VisualRoleWarning] = plain.Foreground(lipgloss.Color("11"))
	styles[VisualRoleError] = plain.Foreground(lipgloss.Color("9")).Bold(true)
	styles[VisualRoleMuted] = plain.Foreground(lipgloss.Color("8")).Faint(true)
	return styles
}

func wrapText(value string, width int) string {
	if width <= 0 {
		return value
	}

	lines := strings.Split(value, "\n")
	for index, line := range lines {
		lines[index] = strings.Join(wrapLine(line, width), "\n")
	}
	return strings.Join(lines, "\n")
}

func wrapLine(line string, width int) []string {
	if lipgloss.Width(line) <= width || strings.TrimSpace(line) == "" {
		return []string{line}
	}

	words := strings.Fields(line)
	lines := make([]string, 0, len(words))
	current := ""
	for _, word := range words {
		if lipgloss.Width(word) > width {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
			lines = append(lines, splitWord(word, width)...)
			continue
		}
		if current == "" {
			current = word
			continue
		}
		if lipgloss.Width(current+" "+word) <= width {
			current += " " + word
			continue
		}
		lines = append(lines, current)
		current = word
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func splitWord(word string, width int) []string {
	runes := []rune(word)
	parts := make([]string, 0, len(runes))
	for len(runes) > 0 {
		count := 1
		for count < len(runes) && lipgloss.Width(string(runes[:count+1])) <= width {
			count++
		}
		parts = append(parts, string(runes[:count]))
		runes = runes[count:]
	}
	return parts
}

func stripTerminalControl(value string) string {
	value = ansi.Strip(value)
	var output strings.Builder
	output.Grow(len(value))
	for _, character := range value {
		switch character {
		case '\n', '\r', '\t':
			output.WriteRune(character)
		default:
			if unicode.IsControl(character) {
				output.WriteRune('\uFFFD')
				continue
			}
			output.WriteRune(character)
		}
	}
	return output.String()
}
