package fork

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"sync"

	"github.com/hackycy/hackycy-cli/internal/logging"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
)

var errGitForkRequiresInteractive = errors.New("git fork requires an interactive terminal")

func runFork(options *Options) error {
	_, err := executeFork(options)
	return err
}

func executeFork(options *Options) (Result, error) {
	if options == nil || options.Context == nil || options.Config == nil || options.WorkingDirectory == nil || options.HTTP == nil || options.Terminal == nil || options.Git == nil {
		return Result{}, errors.New("git fork options are incomplete")
	}
	ctx, cancel := context.WithCancel(options.Context)
	defer cancel()
	run := options.Terminal.Open(ctx)
	defer run.Close()
	adapter := newTerminalGitForkAdapter(run, cancel)

	store, err := options.Config()
	if err != nil {
		return Result{}, err
	}
	provider, err := NewProviderClient(options.HTTP)
	if err != nil {
		return Result{}, err
	}
	module, err := New(Dependencies{
		Config:           store,
		WorkingDirectory: options.WorkingDirectory,
		Directories:      osForkDirectoryReader{},
		Prompter:         adapter,
		Provider:         provider,
		Extractor:        OSArchiveExtractor{},
		CloneRunner:      gitRunnerAdapter{runner: options.Git},
		Remover:          osForkDirectoryRemover{},
		Presenter:        adapter,
		Tracker:          adapter,
	})
	if err != nil {
		return Result{}, err
	}
	result, err := module.Run(ctx, Input{Repository: options.Repository, Destination: options.Destination})
	if err != nil {
		return result, err
	}
	return result, nil
}

type osForkDirectoryReader struct{}

func (osForkDirectoryReader) ReadDir(path string) ([]fs.DirEntry, error) {
	return os.ReadDir(path)
}

type osForkDirectoryRemover struct{}

func (osForkDirectoryRemover) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

type terminalGitForkAdapter struct {
	run           terminalexperience.ExperienceRun
	requestCancel context.CancelFunc
}

func newTerminalGitForkAdapter(run terminalexperience.ExperienceRun, requestCancel context.CancelFunc) *terminalGitForkAdapter {
	return &terminalGitForkAdapter{run: run, requestCancel: requestCancel}
}

func (adapter *terminalGitForkAdapter) ConfirmOverwrite(prompt OverwritePrompt) (bool, bool, error) {
	answer, err := adapter.run.Ask(terminalexperience.InteractionRequest{
		Kind:         terminalexperience.InteractionConfirm,
		Message:      prompt.Message,
		HasDefault:   true,
		Default:      terminalexperience.InteractionAnswer{Confirmed: true},
		CancelValues: []string{"q", "quit", "cancel"},
		PlainPrompt:  prompt.Message + " [Y/n]: ",
		ParsePlain: func(value string) (terminalexperience.InteractionAnswer, error) {
			switch strings.ToLower(strings.TrimSpace(value)) {
			case "y", "yes":
				return terminalexperience.InteractionAnswer{Confirmed: true}, nil
			case "n", "no":
				return terminalexperience.InteractionAnswer{Confirmed: false}, nil
			default:
				return terminalexperience.InteractionAnswer{}, errors.New("Invalid confirmation")
			}
		},
	})
	if errors.Is(err, terminalexperience.ErrInteractionCancelled) || errors.Is(err, context.Canceled) {
		return false, true, nil
	}
	if errors.Is(err, terminalexperience.ErrAutomationInteraction) {
		return false, false, errGitForkRequiresInteractive
	}
	if err != nil {
		return false, false, err
	}
	return answer.Confirmed, false, nil
}

func (adapter *terminalGitForkAdapter) Introduction() {
	_ = adapter.run.Notice(gitForkIntroductionDocument())
}

func (adapter *terminalGitForkAdapter) Cancelled() {
	_ = adapter.run.Result(gitForkCancelledDocument())
}

func (adapter *terminalGitForkAdapter) Outcome(result Result) {
	_ = adapter.run.Result(gitForkOutcomeDocument(result))
}

func (adapter *terminalGitForkAdapter) Start(_ context.Context) (PhaseReporter, error) {
	reporter := &terminalGitForkPhaseReporter{
		updates:  make(chan terminalexperience.OperationPhase, 1),
		finished: make(chan struct{}),
	}
	go func() {
		err := adapter.run.Track(terminalexperience.TrackedOperation{
			Label:         "Git Fork",
			Updates:       reporter.updates,
			RequestCancel: adapter.requestCancel,
		})
		reporter.complete(err)
	}()
	return reporter, nil
}

type terminalGitForkPhaseReporter struct {
	updates  chan terminalexperience.OperationPhase
	finished chan struct{}

	mu     sync.Mutex
	closed bool
	err    error
}

func (reporter *terminalGitForkPhaseReporter) Report(phase Phase) {
	update := terminalGitForkPhase(phase)
	select {
	case reporter.updates <- update:
	case <-reporter.finished:
	}
}

func (reporter *terminalGitForkPhaseReporter) Close() error {
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

func (reporter *terminalGitForkPhaseReporter) complete(err error) {
	reporter.mu.Lock()
	reporter.err = err
	reporter.mu.Unlock()
	close(reporter.finished)
}

func terminalGitForkPhase(phase Phase) terminalexperience.OperationPhase {
	name := "Resolving repository"
	detail := phase.Repository
	switch phase.Kind {
	case PhaseDefaultBranch:
		name = "Fetching default branch"
		detail = phase.Ref
	case PhaseArchive:
		name = "Downloading archive"
		detail = phase.Ref
	case PhaseClone:
		name = "Falling back to git clone"
		detail = phase.Destination
	case PhaseReady:
		name = "Project ready"
		detail = phase.Destination
	}
	return terminalexperience.OperationPhase{Name: name, Detail: detail, State: terminalGitForkPhaseState(phase.State)}
}

func terminalGitForkPhaseState(state PhaseState) terminalexperience.PhaseState {
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

func gitForkIntroductionDocument() terminalexperience.PresentationDocument {
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{
		{Role: terminalexperience.VisualRoleTitle, Text: "HACKYCY CLI"},
		{Role: terminalexperience.VisualRoleActive, Text: "Git Fork"},
	}}
}

func gitForkCancelledDocument() terminalexperience.PresentationDocument {
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleWarning, Text: "Cancelled"}}}
}

func gitForkOutcomeDocument(result Result) terminalexperience.PresentationDocument {
	blocks := []terminalexperience.PresentationBlock{
		gitForkBlock(terminalexperience.VisualRoleMuted, fmt.Sprintf("Resolved: %s/%s/%s (%s)", result.Repository.Host, result.Repository.Owner, result.Repository.Name, result.Repository.ProviderType)),
	}
	if result.Repository.Ref == "" && result.Ref != "" {
		blocks = append(blocks, gitForkBlock(terminalexperience.VisualRoleActive, "Branch: "+result.Ref))
	}
	if result.DefaultBranchError != nil {
		blocks = append(blocks,
			gitForkBlock(terminalexperience.VisualRoleError, "Failed to get default branch: "+logging.Redact(result.DefaultBranchError.Error())),
			gitForkBlock(terminalexperience.VisualRoleWarning, "Falling back to git clone with remote default branch."),
		)
	}
	if result.ArchiveError != nil {
		blocks = append(blocks,
			gitForkBlock(terminalexperience.VisualRoleError, "Archive download failed: "+logging.Redact(result.ArchiveError.Error())),
			gitForkBlock(terminalexperience.VisualRoleWarning, "Falling back to git clone..."),
		)
	}
	switch result.Acquisition {
	case "archive":
		blocks = append(blocks, gitForkBlock(terminalexperience.VisualRoleSuccess, "Archive downloaded and extracted"))
	case "clone":
		if result.ArchiveError == nil {
			blocks = append(blocks, gitForkBlock(terminalexperience.VisualRoleWarning, "Falling back to git clone..."))
		}
		blocks = append(blocks, gitForkBlock(terminalexperience.VisualRoleSuccess, "Cloned and cleaned up"))
	}
	blocks = append(blocks, gitForkBlock(terminalexperience.VisualRoleSuccess, "Done! Project created at "+result.Destination))
	return terminalexperience.PresentationDocument{Blocks: blocks}
}

func gitForkBlock(role terminalexperience.VisualRole, text string) terminalexperience.PresentationBlock {
	return terminalexperience.PresentationBlock{Role: role, Text: text}
}
