package terminal

import (
	"io"
	"strings"

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
		output.WriteString(stripTerminalControl(block.Text))
	}
	if output.Len() > 0 && !strings.HasSuffix(output.String(), "\n") {
		output.WriteByte('\n')
	}
	return output.String()
}

// WritePlain writes a durable Command Result to its stdout destination.
func WritePlain(output io.Writer, document PresentationDocument) error {
	_, err := io.WriteString(output, RenderPlain(document))
	return err
}

// RichOptions controls terminal-owned Rich Interactive presentation behavior.
type RichOptions struct {
	Width int
	Color bool
}

// WriteRich writes a durable Rich Interactive Command Result to stdout.
func WriteRich(output io.Writer, document PresentationDocument, options RichOptions) error {
	_, err := io.WriteString(output, renderRich(document, options))
	return err
}

func renderRich(document PresentationDocument, options RichOptions) string {
	var output strings.Builder
	styles := richStyles(options.Color)
	for index, block := range document.Blocks {
		if index > 0 && output.Len() > 0 && !strings.HasSuffix(output.String(), "\n") {
			output.WriteByte('\n')
		}
		text := wrapText(stripTerminalControl(block.Text), options.Width)
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
	return ansi.Strip(value)
}
