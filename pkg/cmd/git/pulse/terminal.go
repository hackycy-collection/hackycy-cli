package pulse

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
)

var errGitPulseRequiresInteractive = errors.New("git pulse requires an interactive terminal")

func runPulse(options *Options) error {
	if options == nil || options.Context == nil || options.WorkingDirectory == nil || options.Terminal == nil || options.Git == nil || options.Now == nil {
		return errors.New("git pulse options are incomplete")
	}
	ctx, cancel := context.WithCancel(options.Context)
	defer cancel()
	run := options.Terminal.Open(ctx)
	defer run.Close()
	adapter := newTerminalPulseAdapter(run, cancel)
	module, err := New(Dependencies{
		WorkingDirectory: options.WorkingDirectory,
		Stater:           osPathStater{},
		Reader:           osDirectoryReader{},
		Yield:            runtime.Gosched,
		Git:              gitRunnerAdapter{runner: options.Git},
		Prompter:         adapter,
		Presenter:        adapter,
		Tracker:          adapter,
		Now:              options.Now,
	})
	if err != nil {
		return err
	}
	_, err = module.Run(ctx, Input{Directory: options.Directory, Days: options.Days})
	return err
}

type osPathStater struct{}

func (osPathStater) Stat(path string) (fs.FileInfo, error) {
	return os.Stat(path)
}

type osDirectoryReader struct{}

func (osDirectoryReader) ReadDir(path string) ([]os.DirEntry, error) {
	return os.ReadDir(path)
}

type terminalPulseAdapter struct {
	run           terminalexperience.ExperienceRun
	requestCancel context.CancelFunc
}

func newTerminalPulseAdapter(run terminalexperience.ExperienceRun, requestCancel context.CancelFunc) *terminalPulseAdapter {
	return &terminalPulseAdapter{run: run, requestCancel: requestCancel}
}

func (adapter *terminalPulseAdapter) SelectDays(prompt DayPrompt) (int, bool, error) {
	answer, cancelled, err := adapter.ask(pulseDayRequest(prompt))
	if err != nil || cancelled {
		return 0, cancelled, err
	}
	days, err := strconv.Atoi(answer.Value)
	if err != nil {
		return 0, false, err
	}
	return days, false, nil
}

func (adapter *terminalPulseAdapter) SelectAuthors(prompt AuthorPrompt) ([]string, bool, error) {
	answer, cancelled, err := adapter.ask(pulseAuthorRequest(prompt))
	if err != nil || cancelled {
		return nil, cancelled, err
	}
	return append([]string(nil), answer.Values...), false, nil
}

func (adapter *terminalPulseAdapter) Introduction(root string) {
	document := terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{
		{Role: terminalexperience.VisualRoleTitle, Text: "HACKYCY CLI"},
		{Role: terminalexperience.VisualRoleActive, Text: "Git Commit Tree"},
		{Role: terminalexperience.VisualRoleMuted, Text: "Workspace: " + root},
	}}
	_ = adapter.run.Notice(document)
}

func (adapter *terminalPulseAdapter) RepositoriesFound(count int) {
	_ = adapter.run.Notice(pulseDocument(fmt.Sprintf("Found %d %s", count, pulsePlural(count, "repository", "repositories")), terminalexperience.VisualRoleSuccess))
}

func (adapter *terminalPulseAdapter) NoRepositories() {
	_ = adapter.run.Result(pulseDocument("No Git repositories found.", terminalexperience.VisualRoleWarning))
}

func (adapter *terminalPulseAdapter) NoCommits() {
	_ = adapter.run.Result(pulseDocument("No commits found in the specified date range.", terminalexperience.VisualRoleWarning))
}

func (adapter *terminalPulseAdapter) Cancelled() {
	_ = adapter.run.Result(pulseDocument("Operation cancelled.", terminalexperience.VisualRoleError))
}

func (adapter *terminalPulseAdapter) Present(report Report) {
	blocks := []terminalexperience.PresentationBlock{{
		Role: terminalexperience.VisualRoleSuccess,
		Text: fmt.Sprintf("Found %d %s in %d %s", report.CommitCount, pulsePlural(report.CommitCount, "commit", "commits"), len(report.Repositories), pulsePlural(len(report.Repositories), "repository", "repositories")),
	}}
	for _, repository := range report.Repositories {
		blocks = append(blocks,
			terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleActive, Text: fmt.Sprintf("%s (%d %s)", filepath.Base(repository.Path), len(repository.Commits), pulsePlural(len(repository.Commits), "commit", "commits"))},
			terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleMuted, Text: "   " + filepath.Dir(repository.Path) + string(filepath.Separator)},
		)
		for index, commit := range repository.Commits {
			connector := "|-"
			if index == len(repository.Commits)-1 {
				connector = "`-"
			}
			blocks = append(blocks, terminalexperience.PresentationBlock{Role: terminalexperience.VisualRolePlain, Text: fmt.Sprintf("   %s %s | %s | %s", connector, commit.Date, commit.Author, commit.Subject)})
		}
	}
	_ = adapter.run.Result(terminalexperience.PresentationDocument{Blocks: blocks})
}

func (adapter *terminalPulseAdapter) Start(_ context.Context, _ PhaseKind) (PhaseReporter, error) {
	reporter := &terminalPulsePhaseReporter{
		updates:  make(chan terminalexperience.OperationPhase, 1),
		finished: make(chan struct{}),
	}
	go func() {
		err := adapter.run.Track(terminalexperience.TrackedOperation{
			Label:         "Git Pulse",
			Updates:       reporter.updates,
			RequestCancel: adapter.requestCancel,
		})
		reporter.complete(err)
	}()
	return reporter, nil
}

func (adapter *terminalPulseAdapter) ask(request terminalexperience.InteractionRequest) (terminalexperience.InteractionAnswer, bool, error) {
	answer, err := adapter.run.Ask(request)
	if errors.Is(err, terminalexperience.ErrInteractionCancelled) || errors.Is(err, context.Canceled) {
		return terminalexperience.InteractionAnswer{}, true, nil
	}
	if errors.Is(err, terminalexperience.ErrAutomationInteraction) {
		return terminalexperience.InteractionAnswer{}, false, errGitPulseRequiresInteractive
	}
	if err != nil {
		return terminalexperience.InteractionAnswer{}, false, err
	}
	return answer, false, nil
}

func pulseDayRequest(prompt DayPrompt) terminalexperience.InteractionRequest {
	options := make([]terminalexperience.InteractionOption, 0, len(prompt.Options))
	for _, option := range prompt.Options {
		options = append(options, terminalexperience.InteractionOption{Label: option.Label, Value: strconv.Itoa(option.Value)})
	}
	return terminalexperience.InteractionRequest{
		Kind:         terminalexperience.InteractionSelect,
		Message:      prompt.Message,
		PlainLead:    prompt.Message,
		PlainPrompt:  "> ",
		Options:      options,
		CancelValues: []string{"", "q", "quit", "cancel"},
		ParsePlain: func(value string) (terminalexperience.InteractionAnswer, error) {
			index, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil || index < 1 || index > len(options) {
				return terminalexperience.InteractionAnswer{}, errors.New("Invalid selection")
			}
			return terminalexperience.InteractionAnswer{Value: options[index-1].Value}, nil
		},
	}
}

func pulseAuthorRequest(prompt AuthorPrompt) terminalexperience.InteractionRequest {
	options := make([]terminalexperience.InteractionOption, 0, len(prompt.Options))
	for _, option := range prompt.Options {
		options = append(options, terminalexperience.InteractionOption{Label: option.Label, Value: option.Value})
	}
	request := terminalexperience.InteractionRequest{
		Kind:         terminalexperience.InteractionMultiSelect,
		Message:      prompt.Message,
		PlainLead:    prompt.Message,
		PlainPrompt:  "> ",
		Options:      options,
		CancelValues: []string{"q", "quit", "cancel"},
		ParsePlain: func(value string) (terminalexperience.InteractionAnswer, error) {
			indices := strings.FieldsFunc(value, func(character rune) bool {
				return character == ',' || character == ' ' || character == '\t'
			})
			if len(indices) == 0 {
				if prompt.Required {
					return terminalexperience.InteractionAnswer{}, errors.New("At least one author is required.")
				}
				return terminalexperience.InteractionAnswer{Values: []string{}}, nil
			}
			selected := make([]string, 0, len(indices))
			seen := make(map[int]struct{}, len(indices))
			for _, value := range indices {
				index, err := strconv.Atoi(value)
				if err != nil || index < 1 || index > len(options) {
					return terminalexperience.InteractionAnswer{}, errors.New("Invalid selection")
				}
				index--
				if _, duplicate := seen[index]; duplicate {
					continue
				}
				seen[index] = struct{}{}
				selected = append(selected, options[index].Value)
			}
			return terminalexperience.InteractionAnswer{Values: selected}, nil
		},
	}
	if len(prompt.InitialValues) > 0 {
		request.HasDefault = true
		request.Default = terminalexperience.InteractionAnswer{Values: append([]string(nil), prompt.InitialValues...)}
	}
	return request
}

func pulseDocument(text string, role terminalexperience.VisualRole) terminalexperience.PresentationDocument {
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: role, Text: text}}}
}

func pulseReportText(report Report) string {
	var output strings.Builder
	output.WriteByte('\n')
	_, _ = fmt.Fprintf(&output, "Found %d %s in %d %s\n\n", report.CommitCount, pulsePlural(report.CommitCount, "commit", "commits"), len(report.Repositories), pulsePlural(len(report.Repositories), "repository", "repositories"))
	for groupIndex, repository := range report.Repositories {
		_, _ = fmt.Fprintf(&output, "%s (%d %s)\n", filepath.Base(repository.Path), len(repository.Commits), pulsePlural(len(repository.Commits), "commit", "commits"))
		_, _ = fmt.Fprintf(&output, "   %s%c\n", filepath.Dir(repository.Path), filepath.Separator)
		for commitIndex, commit := range repository.Commits {
			connector := "|-"
			if commitIndex == len(repository.Commits)-1 {
				connector = "`-"
			}
			_, _ = fmt.Fprintf(&output, "   %s %s | %s | %s\n", connector, commit.Date, commit.Author, commit.Subject)
		}
		if groupIndex < len(report.Repositories)-1 {
			_, _ = fmt.Fprintln(&output)
		}
	}
	return output.String()
}

type terminalPulsePhaseReporter struct {
	updates  chan terminalexperience.OperationPhase
	finished chan struct{}

	mu     sync.Mutex
	closed bool
	err    error
}

func (reporter *terminalPulsePhaseReporter) Report(phase Phase) {
	update := terminalPulsePhase(phase)
	select {
	case reporter.updates <- update:
	case <-reporter.finished:
	}
}

func (reporter *terminalPulsePhaseReporter) Close() error {
	reporter.mu.Lock()
	if !reporter.closed {
		close(reporter.updates)
		reporter.closed = true
	}
	reporter.mu.Unlock()
	<-reporter.finished
	reporter.mu.Lock()
	defer reporter.mu.Unlock()
	return reporter.err
}

func (reporter *terminalPulsePhaseReporter) complete(err error) {
	reporter.mu.Lock()
	reporter.err = err
	reporter.mu.Unlock()
	close(reporter.finished)
}

func terminalPulsePhase(phase Phase) terminalexperience.OperationPhase {
	name := "Scanning repositories"
	detail := ""
	switch phase.Kind {
	case PhaseFetch:
		name = "Fetching commits"
		if phase.Total > 0 {
			detail = fmt.Sprintf("[%d/%d]", phase.Completed, phase.Total)
		}
		if phase.Repository != "" {
			detail = strings.TrimSpace(detail + " " + pulseRelativePath(phase.Root, phase.Repository))
		}
	default:
		if phase.Repository != "" {
			detail = fmt.Sprintf("[%d] %s", phase.Completed, pulseRelativePath(phase.Root, phase.Repository))
		}
	}
	return terminalexperience.OperationPhase{Name: name, Detail: detail, State: terminalPulsePhaseState(phase.State)}
}

func terminalPulsePhaseState(state PhaseState) terminalexperience.PhaseState {
	switch state {
	case PhaseCompleted:
		return terminalexperience.PhaseCompleted
	case PhaseCancelled:
		return terminalexperience.PhaseCancelled
	case PhaseFailed:
		return terminalexperience.PhaseFailed
	default:
		return terminalexperience.PhaseActive
	}
}

func pulseRelativePath(root, repository string) string {
	relative, err := filepath.Rel(root, repository)
	if err != nil || relative == "" {
		return "."
	}
	return relative
}

func pulsePlural(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}
