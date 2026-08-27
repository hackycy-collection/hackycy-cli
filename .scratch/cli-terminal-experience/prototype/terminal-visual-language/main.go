// Throwaway prototype: three ycy terminal visual directions, selected with --variant.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type direction int

const (
	signal direction = iota
	ledger
	workbench
)

var allDirections = []direction{signal, ledger, workbench}

func (value direction) String() string {
	switch value {
	case signal:
		return "Signal"
	case ledger:
		return "Ledger"
	default:
		return "Workbench"
	}
}

func parseDirection(value string) direction {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "ledger", "b", "2":
		return ledger
	case "workbench", "c", "3":
		return workbench
	default:
		return signal
	}
}

func main() {
	variant := flag.String("variant", "signal", "signal, ledger, or workbench")
	static := flag.Bool("static", false, "print a static snapshot")
	width := flag.Int("width", 100, "snapshot width when --static is set")
	form := flag.String("form", "", "selection, confirm, or secret")
	flag.Parse()

	choice := parseDirection(*variant)
	if *form != "" {
		runForm(choice, *form)
		return
	}

	if *static || !term.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Print(snapshot(choice, *width))
		return
	}

	program := tea.NewProgram(newGallery(choice), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runForm(choice direction, kind string) {
	styles := formTheme(choice)
	switch strings.ToLower(kind) {
	case "selection":
		selected := ""
		form := huh.NewForm(huh.NewGroup(huh.NewSelect[string]().
			Title("Select a branch to publish").
			Description("The chosen item will be used by the next command.").
			Options(huh.NewOption("main  Production branch", "main"), huh.NewOption("release/1.8  Maintenance branch", "release/1.8"), huh.NewOption("feature/cli  Current work", "feature/cli")).
			Value(&selected))).WithTheme(styles)
		_ = form.Run()
		if selected != "" {
			fmt.Println(successLine(choice, "Selected branch: "+selected))
		}
	case "confirm":
		confirmed := false
		form := huh.NewForm(huh.NewGroup(huh.NewConfirm().
			Title("Remove CM profile \"staging\"?").
			Description("This only removes local credentials.").
			Affirmative("Remove").
			Negative("Keep").
			Value(&confirmed))).WithTheme(styles)
		_ = form.Run()
		if confirmed {
			fmt.Println(successLine(choice, "Profile removed"))
		} else {
			fmt.Println(warningLine(choice, "No changes made"))
		}
	case "secret":
		secret := ""
		form := huh.NewForm(huh.NewGroup(huh.NewInput().
			Title("Provider API key").
			Description("Stored locally and never printed.").
			EchoMode(huh.EchoModePassword).
			Value(&secret))).WithTheme(styles)
		_ = form.Run()
		if secret != "" {
			fmt.Println(successLine(choice, "Credential accepted"))
		}
	default:
		fmt.Fprintln(os.Stderr, "unknown form: use selection, confirm, or secret")
		os.Exit(2)
	}
}

type gallery struct {
	direction direction
	width     int
	frame     int
}

func newGallery(choice direction) gallery {
	return gallery{direction: choice, width: 100}
}

func (model gallery) Init() tea.Cmd { return nextFrame() }

func (model gallery) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return model, tea.Quit
		case "left", "h":
			model.direction = allDirections[(int(model.direction)+len(allDirections)-1)%len(allDirections)]
		case "right", "l":
			model.direction = allDirections[(int(model.direction)+1)%len(allDirections)]
		}
	case tea.WindowSizeMsg:
		model.width = msg.Width
	case frameMsg:
		model.frame++
		return model, nextFrame()
	}
	return model, nil
}

func (model gallery) View() string {
	frames := []string{"·", "•", "●", "•"}
	return snapshotWithSpinner(model.direction, model.width, frames[model.frame%len(frames)])
}

type frameMsg struct{}

func nextFrame() tea.Cmd {
	return tea.Tick(300*time.Millisecond, func(time.Time) tea.Msg { return frameMsg{} })
}

func snapshot(choice direction, width int) string {
	return snapshotWithSpinner(choice, width, "...")
}

func snapshotWithSpinner(choice direction, width int, spin string) string {
	switch choice {
	case ledger:
		return ledgerSnapshot(width, spin)
	case workbench:
		return workbenchSnapshot(width, spin)
	default:
		return signalSnapshot(width, spin)
	}
}

func signalSnapshot(width int, spin string) string {
	pageWidth := clamp(width-6, 54, 96)
	styles := visualTheme(signal)
	bar := strings.Repeat("─", pageWidth)
	progress := progressLine(signal, spin)
	content := []string{
		styles.brand.Render("ycy") + "  " + styles.muted.Render("terminal experience prototype"),
		styles.rule.Render(bar),
		styles.heading.Render(directionHeading(signal)),
		styles.muted.Render(wrap(directionDescription(signal), pageWidth)),
		"",
		styles.section.Render("INTERACTION"),
		selectionBlock(signal),
		"",
		confirmationBlock(signal),
		"",
		secretBlock(signal),
		"",
		styles.section.Render("WORK"),
		progress,
		styles.muted.Render("  Fetching commits  14 / 37  packages/api"),
		styles.muted.Render("  Press ctrl+c to cancel"),
		"",
		styles.section.Render("OUTCOMES"),
		outcomeBlock(signal, pageWidth),
		"",
		styles.section.Render("REPORT"),
		compactReport(signal),
		"",
		styles.section.Render("AUTOMATION SESSION"),
		plainAutomationBlock(pageWidth),
		"",
		styles.rule.Render(bar),
		styles.footer.Render("←/h previous   →/l next   q quit") + "  " + styles.chip.Render(signal.String()+"  "+variantPosition(signal)),
	}

	return "\n" + strings.Join(content, "\n") + "\n"
}

func ledgerSnapshot(width int, spin string) string {
	pageWidth := clamp(width-6, 54, 96)
	styles := visualTheme(ledger)
	bar := strings.Repeat("─", pageWidth)
	content := []string{
		styles.brand.Render("ycy") + "  " + styles.muted.Render("interactive transcript"),
		styles.rule.Render(bar),
		styles.heading.Render(directionHeading(ledger)),
		styles.muted.Render(wrap(directionDescription(ledger), pageWidth)),
		"",
		styles.muted.Render("09:41:02") + " " + styles.label.Render("? Select a script"),
		"          " + styles.selecta.Render("› build") + styles.muted.Render("  go build ./cmd/ycy"),
		"          " + styles.body.Render("  test") + styles.muted.Render("   go test ./..."),
		"          " + styles.body.Render("  lint") + styles.muted.Render("   golangci-lint run"),
		"",
		styles.muted.Render("09:41:04") + " " + styles.label.Render("? Remove CM profile \"staging\"?"),
		"          " + styles.selecta.Render("› Remove") + styles.muted.Render("  Keep"),
		"",
		styles.muted.Render("09:41:05") + " " + styles.label.Render("? Provider API key"),
		"          " + styles.body.Render("• • • • • • • • • • • •") + styles.muted.Render("  hidden input"),
		"",
		styles.muted.Render("09:41:06") + " " + progressLine(ledger, spin),
		styles.muted.Render("09:41:09") + " " + styles.body.Render("Fetching commits  14 / 37"),
		"          " + styles.muted.Render("packages/api  ctrl+c cancels"),
		styles.muted.Render("09:41:14") + " " + successLine(ledger, "Archive created"),
		"          " + styles.muted.Render("dist/ycy_1.8.0.zip"),
		styles.muted.Render("09:41:14") + " " + warningLine(ledger, "3 untracked files were skipped"),
		styles.muted.Render("09:41:16") + " " + errorLine(ledger, "Push rejected"),
		"          " + styles.muted.Render("remote contains newer commits"),
		"",
		styles.label.Render("SUMMARY"),
		"  " + styles.body.Render("37 commits / 3 repositories / latest: api 2m ago"),
		styles.rule.Render(bar),
		styles.label.Render("AUTOMATION SESSION"),
		styles.muted.Render(wrap("Same facts, never a transcript or cursor view.", pageWidth)),
		plainAutomationBlock(pageWidth),
		styles.rule.Render(bar),
		styles.footer.Render("←/h previous   →/l next   q quit") + "  " + styles.chip.Render(ledger.String()+"  "+variantPosition(ledger)),
	}
	return "\n" + strings.Join(content, "\n") + "\n"
}

func workbenchSnapshot(width int, spin string) string {
	pageWidth := clamp(width-6, 54, 96)
	styles := visualTheme(workbench)
	bar := strings.Repeat("─", pageWidth)
	leftWidth := 32
	rightWidth := pageWidth - leftWidth - 3
	reportHeading := styles.label.Render("REPORT")
	if pageWidth >= 72 {
		reportHeading += "  " + styles.muted.Render("compact result remains visible after the task ends")
	} else {
		reportHeading += "\n" + styles.muted.Render(wrap("compact result remains visible after the task ends", pageWidth))
	}
	workArea := ""
	if pageWidth >= 80 {
		workArea = lipgloss.JoinHorizontal(lipgloss.Top, workbenchControls(leftWidth), strings.Repeat(" ", 3), workbenchLive(rightWidth, spin))
	} else {
		workArea = workbenchControls(pageWidth) + "\n\n" + workbenchLive(pageWidth, spin)
	}
	content := []string{
		styles.brand.Render("ycy") + "  " + styles.chip.Render("WORKBENCH") + "  " + styles.muted.Render("prototype"),
		styles.rule.Render(bar),
		styles.heading.Render(directionHeading(workbench)),
		styles.muted.Render(wrap(directionDescription(workbench), pageWidth)),
		"",
		workArea,
		"",
		reportHeading,
		compactReport(workbench),
		"",
		styles.label.Render("AUTOMATION SESSION"),
		styles.muted.Render(wrap("No panels, redraws, hidden input, or screen ownership.", pageWidth)),
		plainAutomationBlock(pageWidth),
		styles.rule.Render(bar),
		styles.footer.Render("←/h previous   →/l next   q quit") + "  " + styles.chip.Render(workbench.String()+"  "+variantPosition(workbench)),
	}
	return "\n" + strings.Join(content, "\n") + "\n"
}

func workbenchControls(width int) string {
	styles := visualTheme(workbench)
	rows := []string{
		styles.section.Render("CONTROL"),
		styles.label.Render("SCRIPT"),
		"  " + styles.selecta.Render("› build") + styles.muted.Render("  go build ./cmd/ycy"),
		"    test" + styles.muted.Render("   go test ./..."),
		"    lint" + styles.muted.Render("   golangci-lint run"),
		"",
		styles.label.Render("REMOVE PROFILE"),
		"  " + styles.selecta.Render("● Remove") + styles.muted.Render("  ○ Keep"),
		"",
		styles.label.Render("API KEY"),
		"  " + styles.body.Render("• • • • • • • • • •") + styles.muted.Render("  hidden"),
		"",
		styles.muted.Render("enter confirm  •  esc cancel"),
	}
	return lipgloss.NewStyle().Width(width).Render(strings.Join(rows, "\n"))
}

func workbenchLive(width int, spin string) string {
	styles := visualTheme(workbench)
	rows := []string{
		styles.section.Render("LIVE TASK"),
		"  " + styles.selecta.Render(spin) + " " + styles.body.Render("Fetching commits"),
		"  " + styles.label.Render("14 / 37") + "  " + styles.muted.Render("packages/api"),
		"  " + styles.muted.Render("ctrl+c cancels the task"),
		"",
		styles.section.Render("OUTCOMES"),
		successLine(workbench, "Archive created"),
		"      " + styles.muted.Render("dist/ycy_1.8.0.zip"),
		warningLine(workbench, "3 files skipped"),
		errorLine(workbench, "Push rejected"),
		"      " + styles.muted.Render("remote contains newer commits"),
	}
	return lipgloss.NewStyle().Width(width).Render(strings.Join(rows, "\n"))
}

func directionHeading(choice direction) string {
	switch choice {
	case ledger:
		return "Quiet ledger"
	case workbench:
		return "Dense workbench"
	default:
		return "Clear signal"
	}
}

func directionDescription(choice direction) string {
	switch choice {
	case ledger:
		return "A low-chroma, line-led language for operators who live in scrollback."
	case workbench:
		return "A compact dashboard rhythm for high-information, repeated local work."
	default:
		return "A warm, clear hierarchy that makes decisions and outcomes easy to scan."
	}
}

func variantPosition(choice direction) string {
	return fmt.Sprintf("%d/3", int(choice)+1)
}

type styles struct {
	brand   lipgloss.Style
	heading lipgloss.Style
	section lipgloss.Style
	label   lipgloss.Style
	body    lipgloss.Style
	muted   lipgloss.Style
	rule    lipgloss.Style
	footer  lipgloss.Style
	chip    lipgloss.Style
	success lipgloss.Style
	warn    lipgloss.Style
	failure lipgloss.Style
	selecta lipgloss.Style
}

func visualTheme(choice direction) styles {
	base := styles{
		brand:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69")),
		heading: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")),
		section: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("75")),
		label:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("110")),
		body:    lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		muted:   lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		rule:    lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		footer:  lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		chip:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("16")).Background(lipgloss.Color("75")).Padding(0, 1),
		success: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("78")),
		warn:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")),
		failure: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("204")),
		selecta: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("81")),
	}
	if choice == ledger {
		base.brand = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("250"))
		base.heading = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252"))
		base.section = lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
		base.label = lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
		base.chip = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("244")).Padding(0, 1)
		base.selecta = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252"))
	}
	if choice == workbench {
		base.brand = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("51"))
		base.heading = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("51"))
		base.section = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("141"))
		base.label = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("141"))
		base.chip = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("16")).Background(lipgloss.Color("51")).Padding(0, 1)
		base.selecta = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("51"))
	}
	return base
}

func selectionBlock(choice direction) string {
	styles := visualTheme(choice)
	return strings.Join([]string{
		styles.label.Render("Select a script"),
		"  " + styles.selecta.Render("›") + " " + styles.body.Render("build") + styles.muted.Render("      go build ./cmd/ycy"),
		"    " + styles.body.Render("test") + styles.muted.Render("       go test ./..."),
		"    " + styles.body.Render("lint") + styles.muted.Render("       golangci-lint run"),
		styles.muted.Render("    enter confirm  •  esc cancel"),
	}, "\n")
}

func confirmationBlock(choice direction) string {
	styles := visualTheme(choice)
	return styles.label.Render("Remove CM profile \"staging\"?") + "\n" +
		"  " + styles.selecta.Render("● Remove") + styles.muted.Render("    ○ Keep") + "\n" +
		styles.muted.Render("  Local credentials only.") + "\n" +
		styles.muted.Render("  enter confirm  •  esc cancel")
}

func secretBlock(choice direction) string {
	styles := visualTheme(choice)
	return styles.label.Render("Provider API key") + "\n" +
		"  " + styles.body.Render("• • • • • • • • • • • • • •") + "\n" +
		styles.muted.Render("  Hidden input. Stored locally and never printed.")
}

func progressLine(choice direction, spin string) string {
	styles := visualTheme(choice)
	return "  " + styles.selecta.Render(spin) + " " + styles.body.Render("Fetching commits") + "  " + styles.muted.Render("14 / 37")
}

func successLine(choice direction, message string) string {
	styles := visualTheme(choice)
	return "  " + styles.success.Render("OK") + "   " + styles.body.Render(message)
}

func warningLine(choice direction, message string) string {
	styles := visualTheme(choice)
	return "  " + styles.warn.Render("WARN") + " " + styles.body.Render(message)
}

func errorLine(choice direction, message string) string {
	styles := visualTheme(choice)
	return "  " + styles.failure.Render("ERROR") + " " + styles.body.Render(message)
}

func outcomeBlock(choice direction, width int) string {
	if width < 70 {
		styles := visualTheme(choice)
		return strings.Join([]string{
			successLine(choice, "Archive created"),
			"      " + styles.muted.Render("dist/ycy_1.8.0.zip"),
			warningLine(choice, "3 untracked files were skipped"),
			errorLine(choice, "Push rejected"),
			"      " + styles.muted.Render("remote contains newer commits"),
		}, "\n")
	}
	return strings.Join([]string{
		successLine(choice, "Archive created  dist/ycy_1.8.0.zip"),
		warningLine(choice, "3 untracked files were skipped"),
		errorLine(choice, "Push rejected: remote contains newer commits"),
	}, "\n")
}

func compactReport(choice direction) string {
	styles := visualTheme(choice)
	rows := []string{
		styles.label.Render("Repository") + "                 " + styles.label.Render("Commits") + "  " + styles.label.Render("Last activity"),
		styles.body.Render("api") + strings.Repeat(" ", 24) + styles.body.Render("18") + strings.Repeat(" ", 7) + styles.muted.Render("2m ago"),
		styles.body.Render("worker") + strings.Repeat(" ", 21) + styles.body.Render("11") + strings.Repeat(" ", 7) + styles.muted.Render("18m ago"),
		styles.body.Render("web") + strings.Repeat(" ", 24) + styles.body.Render("8") + strings.Repeat(" ", 8) + styles.muted.Render("1h ago"),
		styles.muted.Render("37 commits across 3 repositories"),
	}
	return strings.Join(rows, "\n")
}

func plainAutomationBlock(width int) string {
	return strings.Join([]string{
		"  $ ycy run",
		automationLine(width, "error: a selection is required; this command cannot prompt in an Automation Session"),
		"",
		"  $ ycy config cm remove staging",
		automationLine(width, "error: confirmation is required; this command cannot prompt in an Automation Session"),
		"",
		"  $ ycy config cm add",
		automationLine(width, "error: a secret is required; ycy never reads secret input from stdin in an Automation Session"),
		"",
		"  $ ycy git pulse .",
		automationLine(width, "progress: fetching commits 14/37 packages/api"),
		automationLine(width, "result: archive created dist/ycy_1.8.0.zip"),
		automationLine(width, "warning: 3 untracked files were skipped"),
		automationLine(width, "error: push rejected; remote contains newer commits"),
		automationLine(width, "report: api 18 2m ago | worker 11 18m ago | web 8 1h ago"),
	}, "\n")
}

func automationLine(width int, text string) string {
	const indent = "  "
	return indent + strings.ReplaceAll(wrap(text, width-len(indent)), "\n", "\n"+indent)
}

func wrap(text string, width int) string {
	if width < 16 {
		return text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	lines := make([]string, 0, len(words))
	line := words[0]
	for _, word := range words[1:] {
		if len([]rune(line))+1+len([]rune(word)) > width {
			lines = append(lines, line)
			line = word
			continue
		}
		line += " " + word
	}
	lines = append(lines, line)
	return strings.Join(lines, "\n")
}

func formTheme(choice direction) *huh.Theme {
	if choice == ledger {
		return huh.ThemeDracula()
	}
	if choice == workbench {
		return huh.ThemeBase16()
	}
	return huh.ThemeCharm()
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
