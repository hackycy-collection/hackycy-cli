package run

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"

	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
)

var errRunRequiresInteractive = errors.New("run requires an interactive terminal")

func runRun(options *Options) error {
	run := options.Terminal.Open(options.Context)
	closed := false
	closeRun := func() error {
		if closed {
			return nil
		}
		closed = true
		return run.Close()
	}
	defer closeRun()

	adapter := newTerminalRunAdapter(run, options.Terminal.Session())
	module, err := New(Dependencies{
		WorkingDirectory: options.WorkingDirectory,
		Reader:           options.Reader,
		Exists:           options.Exists,
		Prompter:         adapter,
		Runner: releasedRunChildRunner{
			release: closeRun,
			runner:  options.Runner,
		},
		Presenter: adapter,
	})
	if err != nil {
		return err
	}
	result, err := module.Run(options.Context, Input{Directory: options.Directory})
	if err != nil {
		return err
	}
	if err := adapter.Flush(); err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return &runChildOutcome{code: result.ExitCode}
	}
	return nil
}

type releasedRunChildRunner struct {
	release func() error
	runner  ChildRunner
}

func (runner releasedRunChildRunner) Run(ctx context.Context, request ChildRequest) (Result, error) {
	if err := runner.release(); err != nil {
		return Result{}, err
	}
	return runner.runner.Run(ctx, request)
}

type osRunFileReader struct{}

func (osRunFileReader) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func osRunPathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

type terminalRunAdapter struct {
	run     terminalexperience.ExperienceRun
	session terminalexperience.Session
	pending []terminalexperience.PresentationDocument
}

func newTerminalRunAdapter(run terminalexperience.ExperienceRun, session terminalexperience.Session) *terminalRunAdapter {
	return &terminalRunAdapter{run: run, session: session}
}

func (adapter *terminalRunAdapter) SelectScript(prompt ScriptPrompt) (string, bool, error) {
	answer, cancelled, err := adapter.ask(terminalexperience.InteractionRequest{
		Kind:         terminalexperience.InteractionSelect,
		Message:      prompt.Message,
		PlainLead:    prompt.Message,
		PlainPrompt:  "> ",
		Options:      runScriptOptions(prompt.Options),
		CancelValues: []string{"", "q", "quit", "cancel"},
		ParsePlain: func(value string) (terminalexperience.InteractionAnswer, error) {
			return parseRunSelection(value, runScriptOptions(prompt.Options))
		},
	})
	return answer.Value, cancelled, err
}

func (adapter *terminalRunAdapter) SelectPackageManager(prompt PackageManagerPrompt) (PackageManager, bool, error) {
	answer, cancelled, err := adapter.ask(terminalexperience.InteractionRequest{
		Kind:         terminalexperience.InteractionSelect,
		Message:      prompt.Message,
		PlainLead:    prompt.Message,
		PlainPrompt:  "> ",
		Options:      runPackageManagerOptions(prompt.Options),
		CancelValues: []string{"", "q", "quit", "cancel"},
		ParsePlain: func(value string) (terminalexperience.InteractionAnswer, error) {
			return parseRunSelection(value, runPackageManagerOptions(prompt.Options))
		},
	})
	return PackageManager(answer.Value), cancelled, err
}

func (adapter *terminalRunAdapter) Intro(message string) {
	if adapter.session.Kind == terminalexperience.RichInteractive {
		adapter.present(terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{
			{Role: terminalexperience.VisualRoleTitle, Text: "HACKYCY CLI"},
			{Role: terminalexperience.VisualRoleActive, Text: message},
		}})
		return
	}
	adapter.present(terminalRunDocument(adapter.session, "HACKYCY CLI\n\n"+message, terminalexperience.VisualRolePlain))
}

func (adapter *terminalRunAdapter) Info(message string) {
	adapter.present(terminalRunDocument(adapter.session, message, terminalexperience.VisualRoleActive))
}

func (adapter *terminalRunAdapter) Blank() {
	adapter.present(terminalRunDocument(adapter.session, "\n", terminalexperience.VisualRolePlain))
}

func (adapter *terminalRunAdapter) Cancel(message string) {
	adapter.present(terminalRunDocument(adapter.session, message, terminalexperience.VisualRoleWarning))
}

func (adapter *terminalRunAdapter) Flush() error {
	if adapter.session.Kind != terminalexperience.Automation {
		return nil
	}
	for _, document := range adapter.pending {
		if err := adapter.run.Present(document); err != nil {
			return err
		}
	}
	adapter.pending = nil
	return nil
}

func (adapter *terminalRunAdapter) ask(request terminalexperience.InteractionRequest) (terminalexperience.InteractionAnswer, bool, error) {
	answer, err := adapter.run.Ask(request)
	if errors.Is(err, terminalexperience.ErrInteractionCancelled) || errors.Is(err, context.Canceled) {
		return terminalexperience.InteractionAnswer{}, true, nil
	}
	if errors.Is(err, terminalexperience.ErrAutomationInteraction) {
		return terminalexperience.InteractionAnswer{}, false, errRunRequiresInteractive
	}
	if err != nil {
		return terminalexperience.InteractionAnswer{}, false, err
	}
	return answer, false, nil
}

func (adapter *terminalRunAdapter) present(document terminalexperience.PresentationDocument) {
	if adapter.session.Kind == terminalexperience.Automation {
		adapter.pending = append(adapter.pending, document)
		return
	}
	_ = adapter.run.Present(document)
}

func terminalRunDocument(session terminalexperience.Session, text string, role terminalexperience.VisualRole) terminalexperience.PresentationDocument {
	if session.Kind != terminalexperience.RichInteractive {
		role = terminalexperience.VisualRolePlain
	}
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: role, Text: text}}}
}

func runScriptOptions(options []ScriptChoice) []terminalexperience.InteractionOption {
	result := make([]terminalexperience.InteractionOption, 0, len(options))
	for _, option := range options {
		result = append(result, terminalexperience.InteractionOption{
			Label:       option.Label,
			Value:       option.Value,
			Description: option.Hint,
		})
	}
	return result
}

func runPackageManagerOptions(options []PackageManagerChoice) []terminalexperience.InteractionOption {
	result := make([]terminalexperience.InteractionOption, 0, len(options))
	for _, option := range options {
		result = append(result, terminalexperience.InteractionOption{Label: option.Label, Value: string(option.Value)})
	}
	return result
}

func parseRunSelection(value string, options []terminalexperience.InteractionOption) (terminalexperience.InteractionAnswer, error) {
	index, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || index < 1 || index > len(options) {
		return terminalexperience.InteractionAnswer{}, errors.New("Invalid selection")
	}
	return terminalexperience.InteractionAnswer{Value: options[index-1].Value}, nil
}
