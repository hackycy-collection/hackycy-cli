package terminal

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestOpenConsoleNormalizesSafeBoundedDescriptorBeforeRichUse(t *testing.T) {
	runtime := NewExperience(ExperienceOptions{Capabilities: Capabilities{Interaction: RichInteractive}})
	run, err := runtime.OpenConsole(context.Background(), ConsoleDescriptor{
		Command:  "  YCY\x1b[31m CONFIG  ",
		Target:   "  profile\x01  ",
		Metadata: []ConsoleMetadata{{Label: " workspace ", Value: " repo\x1b[0m "}},
	})
	if err != nil {
		t.Fatalf("OpenConsole() error = %v", err)
	}

	concrete, ok := run.(*runtimeRun)
	if !ok {
		t.Fatalf("OpenConsole() = %T, want *runtimeRun", run)
	}
	if concrete.console.Command != "YCY CONFIG" || concrete.console.Target != "profile�" || concrete.console.Status != "READY" || len(concrete.console.Metadata) != 1 || concrete.console.Metadata[0] != (ConsoleMetadata{Label: "workspace", Value: "repo"}) {
		t.Fatalf("console descriptor = %#v", concrete.console)
	}
}

func TestOpenConsoleRejectsInvalidDescriptorWithoutOpeningRun(t *testing.T) {
	runtime := NewExperience(ExperienceOptions{})
	_, err := runtime.OpenConsole(context.Background(), ConsoleDescriptor{Command: " "})
	if !errors.Is(err, ErrInvalidConsoleDescriptor) {
		t.Fatalf("OpenConsole() error = %v, want ErrInvalidConsoleDescriptor", err)
	}

	_, err = runtime.OpenConsole(context.Background(), ConsoleDescriptor{
		Command: "YCY",
		Metadata: []ConsoleMetadata{
			{Label: "one", Value: "1"}, {Label: "two", Value: "2"},
			{Label: "three", Value: "3"}, {Label: "four", Value: "4"},
			{Label: "five", Value: "5"},
		},
	})
	if !errors.Is(err, ErrInvalidConsoleDescriptor) {
		t.Fatalf("OpenConsole() metadata error = %v, want ErrInvalidConsoleDescriptor", err)
	}

	_, err = runtime.OpenConsole(context.Background(), ConsoleDescriptor{Command: strings.Repeat("x", maxConsoleField+1)})
	if !errors.Is(err, ErrInvalidConsoleDescriptor) {
		t.Fatalf("OpenConsole() size error = %v, want ErrInvalidConsoleDescriptor", err)
	}
}

func TestOpenUsesTheBCompatibleDefaultConsoleDescriptor(t *testing.T) {
	run := NewExperience(ExperienceOptions{}).Open(context.Background()).(*runtimeRun)
	if run.console.Command != "YCY" || run.console.Target != "terminal session" || run.console.Status != "READY" || len(run.console.Metadata) != 1 {
		t.Fatalf("default descriptor = %#v", run.console)
	}
}

func TestConsoleWideViewKeepsStableShellRegions(t *testing.T) {
	model := newRichRootModelWithConsole(96, 30, false, ConsoleDescriptor{
		Command: "YCY CONFIG",
		Target:  "profile demo",
		Status:  "READY",
		Metadata: []ConsoleMetadata{
			{Label: "workspace", Value: "repo"},
			{Label: "provider", Value: "github"},
		},
	})
	model.mode = richTrackMode
	model.track = &trackedState{label: "sync", phases: []OperationPhase{
		{ID: "scan", Name: "Scan", State: PhaseCompleted, Detail: "repo"},
		{ID: "write", Name: "Write", State: PhaseActive, Detail: "pending"},
	}}
	view := model.View()
	if !view.AltScreen || !view.DisableBracketedPasteMode {
		t.Fatalf("wide view terminal flags = %#v", view)
	}
	for _, needle := range []string{"YCY CONFIG", "profile demo", "workspace repo", "provider github", "STATE", "PHASE", "DETAIL", "✓ DONE", "◆ ACTIVE", "Scan", "Write", "pending"} {
		if !strings.Contains(view.Content, needle) {
			t.Fatalf("wide view missing %q: %q", needle, view.Content)
		}
	}
	if strings.Contains(view.Content, "[done]") || strings.Contains(view.Content, "[active]") {
		t.Fatalf("wide view retained generic phase prefixes: %q", view.Content)
	}
}

func TestConsoleCompactSurfaceUsesTheBStatusHeading(t *testing.T) {
	model := newRichRootModelWithConsole(69, 30, false, defaultConsoleDescriptor())
	model.mode = richTrackMode
	model.track = &trackedState{label: "work", phases: []OperationPhase{{Name: "Phase", State: PhaseActive}}}
	view := model.View().Content
	if !strings.Contains(view, "STATE / PHASE / DETAIL") || !strings.Contains(view, "◆ ACTIVE · Phase") {
		t.Fatalf("compact surface omitted B status structure: %q", view)
	}
}

func TestConsoleCompactViewRetainsOrderedRowsAndActiveRegion(t *testing.T) {
	model := newRichRootModelWithConsole(48, 16, false, ConsoleDescriptor{
		Command:  "YCY GIT",
		Target:   "pulse",
		Metadata: []ConsoleMetadata{{Label: "scope", Value: "workspace"}},
	})
	model.mode = richTrackMode
	model.track = &trackedState{label: "Git Pulse", phases: []OperationPhase{
		{ID: "scan", Name: "Scan", State: PhaseCompleted, Detail: "2 repos"},
		{ID: "fetch", Name: "Fetch", State: PhaseActive, Detail: "commits"},
	}}
	view := model.View().Content
	for _, needle := range []string{"YCY GIT", "scope workspace", "STATE / PHASE / DETAIL", "✓ DONE · Scan · 2 repos", "◆ ACTIVE · Fetch · commits", "Fetch", "commits"} {
		if !strings.Contains(view, needle) {
			t.Fatalf("compact view missing %q: %q", needle, view)
		}
	}
}

func TestConsoleNormalizedProjectionKeepsMetadataSingleLineAndWithinWidth(t *testing.T) {
	longValue := strings.Repeat("workspace-value ", 8) + "\nwith another line"
	runtime := NewExperience(ExperienceOptions{})
	run, err := runtime.OpenConsole(context.Background(), ConsoleDescriptor{
		Command:  "  YCY\tCONFIG\n",
		Target:   " profile\n demo ",
		Metadata: []ConsoleMetadata{{Label: "workspace\nname", Value: longValue}},
	})
	if err != nil {
		t.Fatalf("OpenConsole() error = %v", err)
	}
	concrete := run.(*runtimeRun)
	if concrete.console.Command != "YCY CONFIG" || concrete.console.Target != "profile demo" || concrete.console.Metadata[0].Label != "workspace name" || strings.Contains(concrete.console.Metadata[0].Value, "\n") {
		t.Fatalf("normalized console fields = %#v", concrete.console)
	}

	model := newRichRootModelWithConsole(70, 20, false, concrete.console)
	model.mode = richTrackMode
	model.track = &trackedState{label: "Work", phases: []OperationPhase{{Name: "Phase", State: PhaseActive, Detail: "detail"}}}
	view := model.View().Content
	for _, line := range strings.Split(view, "\n") {
		if lipgloss.Width(line) > 70 {
			t.Fatalf("console line exceeds terminal width: %d > 70: %q", lipgloss.Width(line), line)
		}
	}
	metadata := model.consoleMetadataView(richStyles(false), 20)
	if strings.Contains(metadata, "\n") || lipgloss.Width(metadata) > 20 {
		t.Fatalf("metadata projection is not bounded single-line: %q (width %d)", metadata, lipgloss.Width(metadata))
	}
}

func TestConsoleModelUsesBPaletteAndNoColorRemovesSGR(t *testing.T) {
	console := ConsoleDescriptor{
		Command:  "YCY CONFIG",
		Target:   "profile demo",
		Metadata: []ConsoleMetadata{{Label: "workspace", Value: "repo"}},
	}
	colored := newRichRootModelWithConsole(96, 30, true, console)
	colored.mode = richTrackMode
	colored.track = &trackedState{label: "Work", phases: []OperationPhase{
		{Name: "Done phase", State: PhaseCompleted, Detail: "saved"},
		{Name: "Active phase", State: PhaseActive, Detail: "working"},
	}}
	coloredView := colored.View().Content
	for _, colorCode := range []string{"38;2;255;180;84", "38;2;90;247;142"} {
		if !strings.Contains(coloredView, colorCode) {
			t.Fatalf("colored B view missing palette code %q: %q", colorCode, coloredView)
		}
	}
	if strings.Contains(ansi.Strip(coloredView), "[done]") || strings.Contains(ansi.Strip(coloredView), "[active]") {
		t.Fatalf("colored B view retained generic state prefix: %q", coloredView)
	}

	plain := newRichRootModelWithConsole(96, 30, false, console)
	plain.mode = richTrackMode
	plain.track = colored.track
	plainView := plain.View().Content
	if strings.Contains(plainView, "\x1b[") {
		t.Fatalf("NO_COLOR B view contains SGR/control styling: %q", plainView)
	}
	for _, text := range []string{"STATE", "Done phase", "Active phase", "✓ DONE", "◆ ACTIVE"} {
		if !strings.Contains(plainView, text) {
			t.Fatalf("NO_COLOR B view missing %q: %q", text, plainView)
		}
	}
}

func TestConsoleNoticeStaysAsLatestBoundedActiveContextBelowTable(t *testing.T) {
	model := newRichRootModelWithConsole(96, 30, false, defaultConsoleDescriptor())
	model.mode = richTrackMode
	model.track = &trackedState{label: "Work", phases: []OperationPhase{{Name: "Phase", State: PhaseActive, Detail: "working"}}}
	model.notices = []PresentationDocument{
		{Blocks: []PresentationBlock{{Text: "old context"}}},
		{Blocks: []PresentationBlock{{Text: "latest context"}}},
	}
	view := model.View().Content
	if !strings.Contains(view, "latest context") || strings.Contains(view, "old context") {
		t.Fatalf("notice context = %q", view)
	}
	stateRow := strings.Index(view, "◆ ACTIVE    Phase")
	active := strings.LastIndex(view, "\n   Work")
	context := strings.Index(view, "latest context")
	if stateRow < 0 || active < 0 || context < stateRow || context > active {
		t.Fatalf("notice context displaced table or active region: %q", view)
	}
}

func TestBThemeUsesBottomFocusAndApprovedPalette(t *testing.T) {
	theme := bHuhTheme(true).Theme(true)
	if !theme.Focused.Base.GetBorderBottom() || theme.Focused.Base.GetBorderLeft() {
		t.Fatalf("focused Huh border = bottom:%t left:%t; want bottom-only", theme.Focused.Base.GetBorderBottom(), theme.Focused.Base.GetBorderLeft())
	}
	if got := ansi.Strip(theme.Focused.SelectSelector.String()); got != "◆ " {
		t.Fatalf("select selector = %q, want paired active symbol", got)
	}
	if got := ansi.Strip(theme.Focused.SelectedPrefix.String()); got != "✓ " {
		t.Fatalf("selected prefix = %q, want paired success symbol", got)
	}
	if got := ansi.Strip(theme.Focused.UnselectedPrefix.String()); got != "○ " {
		t.Fatalf("unselected prefix = %q, want paired pending symbol", got)
	}
	if got := theme.Focused.Title.GetForeground(); got == nil {
		t.Fatal("focused title has no B primary color")
	}
}

func TestBThemeCanRenderWithoutColor(t *testing.T) {
	theme := bHuhTheme(false).Theme(true)
	for _, style := range []struct {
		name  string
		value string
	}{
		{name: "title", value: theme.Focused.Title.Render("title")},
		{name: "selected", value: theme.Focused.SelectedPrefix.Render("done")},
	} {
		if strings.Contains(style.value, "\x1b[") {
			t.Fatalf("%s style contains ANSI in no-color mode: %q", style.name, style.value)
		}
	}
}

func TestConsoleFormRowsRetainReachedOrderAndRedactedStepDetail(t *testing.T) {
	model := newRichRootModelWithConsole(96, 30, false, defaultConsoleDescriptor())
	response := make(chan richAskResult, 2)
	show := func(id uint64, request InteractionRequest) {
		_, _ = model.Update(richShowFormMsg{
			id:       id,
			form:     consoleTestForm{},
			answer:   func() InteractionAnswer { return InteractionAnswer{} },
			step:     newConsoleFormStep(id, request),
			response: response,
			ack:      make(chan struct{}),
		})
	}

	show(1, InteractionRequest{Kind: InteractionText, Message: "Workspace", TranscriptLabel: "Workspace"})
	_, _ = model.Update(richFormSubmittedMsg{id: 1})
	<-response
	show(2, InteractionRequest{Kind: InteractionSecret, Message: "Access token", TranscriptLabel: "Access token"})

	view := model.View().Content
	completed := strings.Index(view, "✓ DONE")
	workspace := strings.Index(view, "Workspace")
	active := strings.Index(view, "◆ ACTIVE")
	token := strings.Index(view, "Access token")
	if completed < 0 || workspace < completed || active < workspace || token < active || !strings.Contains(view, "answer captured") || !strings.Contains(view, "redacted input") {
		t.Fatalf("form rows = %q", view)
	}

	_, _ = model.Update(richFormCancelledMsg{id: 2})
	<-response
	view = model.View().Content
	if !strings.Contains(view, "⊘ CANCELLED") || !strings.Contains(view, "Access token") || !strings.Contains(view, "cancelled") {
		t.Fatalf("cancelled form row = %q", view)
	}
}

func TestConsoleTrackRowsRetainCatalogOrderAndFinalDetailsAfterActiveRegionChanges(t *testing.T) {
	model := newRichRootModelWithConsole(96, 30, false, defaultConsoleDescriptor())
	start := func() {
		_, _ = model.Update(richStartTrackMsg{
			label: "Profile setup",
			phases: []OperationPhase{
				{ID: "validate", Name: "Validate source", State: PhasePending},
				{ID: "write", Name: "Write profile", State: PhasePending},
			},
			requestCancel: func() error { return nil },
			ack:           make(chan struct{}),
		})
	}
	update := func(phase OperationPhase) {
		_, _ = model.Update(richTrackPhaseMsg{phase: phase, ack: make(chan struct{})})
	}

	start()
	update(OperationPhase{ID: "validate", State: PhaseActive, Detail: "reading configuration"})
	update(OperationPhase{ID: "validate", State: PhaseCompleted, Detail: "configuration validated"})
	update(OperationPhase{ID: "write", State: PhaseActive, Detail: "persisting profile"})
	update(OperationPhase{ID: "write", State: PhaseCompleted, Detail: "profile persisted"})
	_, _ = model.Update(richFinishTrackMsg{ack: make(chan struct{})})
	_, _ = model.Update(richNoticeMsg{
		document: PresentationDocument{Blocks: []PresentationBlock{{Text: "Follow-up context"}}},
		ack:      make(chan struct{}),
	})

	view := model.View().Content
	if got := strings.Count(view, "Validate source"); got != 1 {
		t.Fatalf("Validate source count = %d, want 1: %q", got, view)
	}
	if got := strings.Count(view, "Write profile"); got != 1 {
		t.Fatalf("Write profile count = %d, want 1: %q", got, view)
	}
	validate := strings.Index(view, "Validate source")
	write := strings.Index(view, "Write profile")
	if validate < 0 || write < validate || !strings.Contains(view, "configuration validated") || !strings.Contains(view, "profile persisted") {
		t.Fatalf("track catalog rows lost order or final detail: %q", view)
	}
	if !strings.Contains(view, "✓ DONE") || !strings.Contains(view, "Follow-up context") {
		t.Fatalf("track table or replacement active region missing: %q", view)
	}
}

type consoleTestForm struct{}

func (consoleTestForm) Init() tea.Cmd { return nil }

func (form consoleTestForm) Update(tea.Msg) (tea.Model, tea.Cmd) { return form, nil }

func (consoleTestForm) View() tea.View { return tea.NewView("active input") }

func (consoleTestForm) configure(int, int, bool) {}

func (consoleTestForm) handlesEscape() bool { return false }
