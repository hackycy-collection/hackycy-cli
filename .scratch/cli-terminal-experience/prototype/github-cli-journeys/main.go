// Throwaway prototype: command-local terminal journeys inspired by GitHub CLI.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type palette struct {
	body    lipgloss.Style
	strong  lipgloss.Style
	muted   lipgloss.Style
	focus   lipgloss.Style
	success lipgloss.Style
	warning lipgloss.Style
	failure lipgloss.Style
	path    lipgloss.Style
}

func newPalette() palette {
	return palette{
		body:    lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		strong:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")),
		muted:   lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		focus:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")),
		success: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42")),
		warning: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220")),
		failure: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("203")),
		path:    lipgloss.NewStyle().Foreground(lipgloss.Color("75")),
	}
}

func main() {
	journey := flag.String("journey", "git-cm", "git-cm, git-pulse, tunnel-connect, tunnel-server, config-cm-add, automation, or all")
	form := flag.String("form", "", "git-cm, git-pulse, tunnel-connect, or config-cm-add")
	width := flag.Int("width", 96, "preview width")
	flag.Parse()

	if *form != "" {
		if !interactiveTerminal() {
			fmt.Fprintln(os.Stderr, "error: form demonstrations require an interactive terminal")
			os.Exit(2)
		}
		runForm(*form)
		return
	}

	preview, ok := renderJourney(*journey, *width)
	if !ok {
		fmt.Fprintln(os.Stderr, "error: unknown journey; use git-cm, git-pulse, tunnel-connect, tunnel-server, config-cm-add, automation, or all")
		os.Exit(2)
	}
	fmt.Print(preview)
}

func interactiveTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) &&
		term.IsTerminal(int(os.Stdout.Fd())) &&
		term.IsTerminal(int(os.Stderr.Fd()))
}

func renderJourney(name string, width int) (string, bool) {
	width = clamp(width, 54, 96)
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "git-cm", "cm":
		return gitCMJourney(width), true
	case "git-pulse", "pulse":
		return gitPulseJourney(width), true
	case "tunnel-connect", "connect":
		return tunnelConnectJourney(width), true
	case "tunnel-server", "server":
		return tunnelServerJourney(width), true
	case "config-cm-add", "cm-add":
		return configCMAddJourney(width), true
	case "automation":
		return automationJourney(width), true
	case "all":
		return allJourneys(width), true
	default:
		return "", false
	}
}

func gitCMJourney(width int) string {
	p := newPalette()
	lines := []string{
		question(p, "Select files to stage"),
	}
	lines = append(lines, gitCMFiles(p, width)...)
	lines = append(lines,
		"  "+p.muted.Render("up/down move  space select  enter confirm  esc cancel"),
		"",
		success(p, "Staged 2 files"),
		working(p, "Generating a commit message with profile work"),
		success(p, "Commit message generated"),
		"",
		p.strong.Render("feat(cli): clarify terminal output"),
		"",
		metadata(p, "Profile", "work / gpt-5"),
		metadata(p, "Context", "2 files  ·  18 facts  ·  3,142 prompt tokens"),
		question(p, "Create this commit? "+p.muted.Render("(Y/n)")),
		success(p, "Commit created "+p.path.Render("8c2a4fd")),
	)
	return finish(lines, width)
}

func gitPulseJourney(width int) string {
	p := newPalette()
	lines := []string{
		question(p, "Select date range"),
		"  " + p.focus.Render(">") + " " + p.strong.Render("Last 7 days"),
		"  " + p.muted.Render("Today"),
		"  " + p.muted.Render("Last 30 days"),
		"",
		working(p, "Scanning workspace"),
		success(p, "Found 3 repositories"),
		working(p, "Fetching commits  14/37  "+p.path.Render("packages/api")),
		success(p, "Collected 37 commits"),
		"",
		p.strong.Render("Last 7 days") + "  " + p.muted.Render("/Users/me/workspace"),
		p.strong.Render("37 commits") + "  " + p.muted.Render("3 repositories  ·  4 authors"),
		"",
		p.muted.Render("ACTIVITY"),
	}
	lines = append(lines, pulseActivity(p, width)...)
	lines = append(lines,
		"",
		p.muted.Render("REPOSITORY")+"  "+p.muted.Render("COMMITS  LAST COMMIT"),
		repositoryRow(p, "api", "18", "2m ago"),
		repositoryRow(p, "worker", "11", "18m ago"),
		repositoryRow(p, "web", "8", "1h ago"),
	)
	return finish(lines, width)
}

func tunnelConnectJourney(width int) string {
	p := newPalette()
	lines := []string{
		question(p, "Select a saved tunnel connection"),
		"  " + p.focus.Render(">") + " " + p.strong.Render("dev.example.net") + "      " + p.muted.Render("token: 8h29********c4bf"),
		"    " + p.body.Render("staging.example.net") + "  " + p.muted.Render("token: 4q7a********e9d1"),
		"    " + p.body.Render("prod.example.net") + "     " + p.muted.Render("token: 9m02********d6ae"),
		"  " + p.muted.Render("enter connect  esc cancel"),
		"",
		success(p, "Connected to "+p.strong.Render("dev.example.net")),
		metadata(p, "Forwarding", p.path.Render("http://127.0.0.1:3000")),
		p.muted.Render("Press ctrl+c to disconnect."),
	}
	return finish(lines, width)
}

func tunnelServerJourney(width int) string {
	p := newPalette()
	lines := []string{
		success(p, "Tunnel server listening on "+p.path.Render(":7000")),
		p.muted.Render("Press ctrl+c to stop."),
		"",
		logLine(p, width, "14:06:21", "INFO", "tunnel.server", "client authenticated", "client=1c9a…4b3e"),
		logLine(p, width, "14:06:22", "INFO", "tunnel.server", "proxy registered", "name=docs"),
		logLine(p, width, "14:13:08", "WARN", "tunnel.server", "client disconnected", "reason=unexpected EOF"),
		logLine(p, width, "14:13:10", "INFO", "tunnel.server", "client reconnected", "client=1c9a…4b3e"),
		logLine(p, width, "14:17:43", "ERROR", "tunnel.server", "proxy rejected", "reason=port already allocated"),
	}
	return finish(lines, width)
}

func configCMAddJourney(width int) string {
	p := newPalette()
	lines := []string{
		question(p, "Add a CM profile"),
		formValue(p, "Profile name", "work"),
		formValue(p, "Provider", "OpenAI-compatible"),
		formValue(p, "Base URL", "https://provider.example/v1"),
		formValue(p, "Model", "gpt-5"),
		formValue(p, "API key", "••••••••••••••••"),
		"",
		success(p, "Saved CM profile "+p.strong.Render("work")),
		metadata(p, "Provider", "OpenAI-compatible  ·  Model gpt-5"),
		p.muted.Render("The API key is stored locally and never printed."),
	}
	return finish(lines, width)
}

func automationJourney(width int) string {
	lines := []string{
		"$ ycy git cm",
		"error: file selection is required; ycy cannot prompt in an Automation Session",
		"",
		"$ ycy git pulse .",
		"progress: scanning workspace",
		"progress: fetching commits 14/37 packages/api",
		"result: 37 commits across 3 repositories",
		"",
		"$ ycy tunnel server",
		"2026-08-26T14:06:21Z INFO  tunnel.server listening address=:7000",
		"2026-08-26T14:06:22Z INFO  tunnel.server client authenticated client=1c9a…4b3e",
		"",
		"$ ycy config cm add",
		"error: a secret is required; ycy never reads secret input from stdin in an Automation Session",
	}
	return plainFinish(lines, width)
}

func allJourneys(width int) string {
	sections := []string{
		gitCMJourney(width),
		gitPulseJourney(width),
		tunnelConnectJourney(width),
		tunnelServerJourney(width),
		configCMAddJourney(width),
		automationJourney(width),
	}
	return strings.Join(sections, "\n"+strings.Repeat("-", clamp(width, 54, 96))+"\n")
}

func question(p palette, text string) string {
	return p.focus.Render("?") + " " + p.strong.Render(text)
}

func gitCMFiles(p palette, width int) []string {
	if width >= 72 {
		return []string{
			selectedFile(p, "cmd/ycy/gitcm.go", "modified"),
			selectedFile(p, "internal/commands/git/cm/run.go", "modified"),
			unselectedFile(p, "README.md", "modified"),
		}
	}
	return []string{
		"  " + p.success.Render("✓") + " " + p.path.Render("cmd/ycy/gitcm.go"),
		"    " + p.muted.Render("modified"),
		"  " + p.success.Render("✓") + " " + p.path.Render("internal/commands/git/cm/run.go"),
		"    " + p.muted.Render("modified"),
		"    " + p.body.Render("README.md"),
		"    " + p.muted.Render("modified"),
	}
}

func selectedFile(p palette, path, state string) string {
	return "  " + p.success.Render("✓") + " " + p.path.Render(pad(path, 38)) + " " + p.muted.Render(state)
}

func unselectedFile(p palette, path, state string) string {
	return "    " + p.body.Render(pad(path, 38)) + " " + p.muted.Render(state)
}

func working(p palette, message string) string {
	return p.focus.Render("⠋") + " " + p.body.Render(message)
}

func success(p palette, message string) string {
	return p.success.Render("✓") + " " + p.body.Render(message)
}

func metadata(p palette, label, value string) string {
	return "  " + p.muted.Render(pad(label, 11)) + " " + p.body.Render(value)
}

func formValue(p palette, label, value string) string {
	return "  " + p.muted.Render(pad(label, 14)) + " " + p.body.Render(value)
}

func repositoryRow(p palette, repository, commits, latest string) string {
	return "  " + p.path.Render(pad(repository, 21)) + p.body.Render(pad(commits, 9)) + p.muted.Render(latest)
}

func pulseActivity(p palette, width int) []string {
	if width >= 72 {
		return []string{
			"  " + p.muted.Render("Mon  Tue  Wed  Thu  Fri  Sat  Sun"),
			"  " + activityCell(p, "▄") + "    " + activityCell(p, "█") + "    " + activityCell(p, "▃") + "    " + activityCell(p, "▆") + "    " + activityCell(p, "▂") + "    " + activityCell(p, "▅") + "    " + activityCell(p, "▁"),
			"  " + p.muted.Render(" 4    9    3    7    2    6    1"),
		}
	}
	return []string{
		activityRow(p, "Mon", 4),
		activityRow(p, "Tue", 9),
		activityRow(p, "Wed", 3),
		activityRow(p, "Thu", 7),
		activityRow(p, "Fri", 2),
		activityRow(p, "Sat", 6),
		activityRow(p, "Sun", 1),
	}
}

func activityCell(p palette, glyph string) string {
	return p.success.Render(glyph)
}

func activityRow(p palette, day string, commits int) string {
	return "  " + p.muted.Render(day) + "  " + p.success.Render(strings.Repeat("█", commits)) + " " + p.muted.Render(fmt.Sprintf("%d", commits))
}

func logLine(p palette, width int, timestamp, level, scope, message, detail string) string {
	levelStyle := p.focus
	switch level {
	case "WARN":
		levelStyle = p.warning
	case "ERROR":
		levelStyle = p.failure
	}
	if width < 72 {
		return p.muted.Render(timestamp) + " " + levelStyle.Render(level) + " " + p.muted.Render(scope) + " " + p.body.Render(message) + "\n  " + p.muted.Render(detail)
	}
	return p.muted.Render(timestamp) + "  " + levelStyle.Render(pad(level, 5)) + " " + p.muted.Render(pad(scope, 15)) + " " + p.body.Render(message) + " " + p.muted.Render(detail)
}

func finish(lines []string, width int) string {
	return "\n" + joinWrapped(lines, width) + "\n"
}

func plainFinish(lines []string, width int) string {
	return "\n" + joinWrapped(lines, width) + "\n"
}

func joinWrapped(lines []string, width int) string {
	wrapped := make([]string, 0, len(lines))
	for _, line := range lines {
		for _, physicalLine := range strings.Split(line, "\n") {
			wrapped = append(wrapped, wrap(physicalLine, width))
		}
	}
	return strings.Join(wrapped, "\n")
}

func wrap(text string, width int) string {
	if width < 16 || len([]rune(text)) <= width {
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
	return strings.Join(append(lines, line), "\n")
}

func pad(value string, width int) string {
	missing := width - len([]rune(value))
	if missing <= 0 {
		return value
	}
	return value + strings.Repeat(" ", missing)
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

func runForm(name string) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "git-cm", "cm":
		runGitCMForm()
	case "git-pulse", "pulse":
		runGitPulseForm()
	case "tunnel-connect", "connect":
		runTunnelConnectForm()
	case "config-cm-add", "cm-add":
		runConfigCMAddForm()
	default:
		fmt.Fprintln(os.Stderr, "error: unknown form; use git-cm, git-pulse, tunnel-connect, or config-cm-add")
		os.Exit(2)
	}
}

func runGitCMForm() {
	selected := []string{"cmd/ycy/gitcm.go", "internal/commands/git/cm/run.go"}
	commit := true
	form := huh.NewForm(
		huh.NewGroup(huh.NewMultiSelect[string]().
			Title("Select files to stage").
			Description("Choose the changes that belong in this commit.").
			Options(
				huh.NewOption("cmd/ycy/gitcm.go  modified", "cmd/ycy/gitcm.go"),
				huh.NewOption("internal/commands/git/cm/run.go  modified", "internal/commands/git/cm/run.go"),
				huh.NewOption("README.md  modified", "README.md"),
			).
			Value(&selected)),
		huh.NewGroup(huh.NewConfirm().
			Title("Create the commit after generating a message?").
			Affirmative("Create commit").
			Negative("Generate only").
			Value(&commit)),
	).WithTheme(githubTheme())
	if form.Run() == nil {
		fmt.Print(gitCMJourney(96))
	}
}

func runGitPulseForm() {
	days := 7
	authors := []string{"Ben", "Maya"}
	form := huh.NewForm(
		huh.NewGroup(huh.NewSelect[int]().
			Title("Select date range").
			Options(huh.NewOption("Today", 1), huh.NewOption("Last 3 days", 3), huh.NewOption("Last 7 days", 7), huh.NewOption("Last 30 days", 30)).
			Value(&days)),
		huh.NewGroup(huh.NewMultiSelect[string]().
			Title("Select authors").
			Description("Leave the default selection to include the active authors.").
			Options(huh.NewOption("Ben", "Ben"), huh.NewOption("Maya", "Maya"), huh.NewOption("Sam", "Sam"), huh.NewOption("Yuki", "Yuki")).
			Value(&authors)),
	).WithTheme(githubTheme())
	if form.Run() == nil {
		fmt.Print(gitPulseJourney(96))
	}
}

func runTunnelConnectForm() {
	connection := "dev"
	form := huh.NewForm(huh.NewGroup(huh.NewSelect[string]().
		Title("Select a saved tunnel connection").
		Options(
			huh.NewOption("dev.example.net      token: 8h29********c4bf", "dev"),
			huh.NewOption("staging.example.net  token: 4q7a********e9d1", "staging"),
			huh.NewOption("prod.example.net     token: 9m02********d6ae", "prod"),
		).
		Value(&connection))).WithTheme(githubTheme())
	if form.Run() == nil {
		fmt.Print(tunnelConnectJourney(96))
	}
}

func runConfigCMAddForm() {
	name := "work"
	provider := "OpenAI-compatible"
	baseURL := "https://provider.example/v1"
	model := "gpt-5"
	apiKey := ""
	form := huh.NewForm(
		huh.NewGroup(huh.NewInput().
			Title("Profile name").
			Description("Used by git cm and other CM-enabled commands.").
			Value(&name)).
			Title("Add a CM profile"),
		huh.NewGroup(huh.NewSelect[string]().
			Title("Provider").
			Options(huh.NewOption("OpenAI-compatible", "OpenAI-compatible"), huh.NewOption("Anthropic-compatible", "Anthropic-compatible")).
			Value(&provider)),
		huh.NewGroup(
			huh.NewInput().Title("Base URL").Value(&baseURL),
			huh.NewInput().Title("Model").Value(&model),
			huh.NewInput().Title("API key").Description("Hidden input. Never printed.").EchoMode(huh.EchoModePassword).Value(&apiKey),
		),
	).WithTheme(githubTheme())
	if form.Run() == nil && apiKey != "" {
		fmt.Print(configCMAddJourney(96))
	}
}

func githubTheme() *huh.Theme {
	p := newPalette()
	theme := huh.ThemeBase()
	theme.Form.Base = lipgloss.NewStyle()
	theme.Group.Title = p.strong
	theme.Group.Description = p.muted
	theme.Focused.Title = p.strong
	theme.Focused.Description = p.muted
	theme.Focused.SelectSelector = p.focus
	theme.Focused.Option = p.body
	theme.Focused.MultiSelectSelector = p.focus
	theme.Focused.SelectedOption = p.body
	theme.Focused.SelectedPrefix = p.success
	theme.Focused.UnselectedOption = p.body
	theme.Focused.UnselectedPrefix = p.muted
	theme.Focused.TextInput.Prompt = p.focus
	theme.Focused.TextInput.Text = p.body
	theme.Focused.TextInput.Placeholder = p.muted
	theme.Focused.TextInput.Cursor = p.focus
	theme.Focused.FocusedButton = p.focus
	theme.Focused.BlurredButton = p.muted
	theme.Focused.ErrorIndicator = p.failure
	theme.Focused.ErrorMessage = p.failure
	theme.Blurred = theme.Focused
	return theme
}
