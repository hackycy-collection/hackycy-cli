package env

import (
	"context"
	"errors"

	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
)

var errExportEnvRequiresInteractive = errors.New("export env requires an interactive terminal")

func runEnv(options *Options) error {
	run := options.Terminal.Open(options.Context)
	defer run.Close()
	adapter := newTerminalExportEnvAdapter(run, options.Terminal.Session())
	module, err := New(Dependencies{
		WorkingDirectory: options.WorkingDirectory,
		Selector:         adapter,
		Reader:           options.Reader,
		Writer:           options.Writer,
		Presenter:        adapter,
	})
	if err != nil {
		return err
	}
	_, err = module.Run(options.Context, Input{
		Directory:   options.Directory,
		Environment: options.Environment,
		Merge:       options.Merge,
		Output:      options.Output,
	})
	return err
}

type terminalExportEnvAdapter struct {
	run     terminalexperience.ExperienceRun
	session terminalexperience.Session
}

func newTerminalExportEnvAdapter(run terminalexperience.ExperienceRun, session terminalexperience.Session) *terminalExportEnvAdapter {
	return &terminalExportEnvAdapter{run: run, session: session}
}

func (adapter *terminalExportEnvAdapter) SelectEnvironment(message string, choices []EnvironmentChoice) (string, bool, error) {
	if adapter.session.Kind == terminalexperience.Automation && len(choices) == 1 {
		return choices[0].Value, false, nil
	}
	answer, err := adapter.run.Ask(terminalexperience.InteractionRequest{
		Kind:         terminalexperience.InteractionSelect,
		Message:      message,
		Options:      exportEnvInteractionOptions(choices),
		CancelValues: []string{"", "q", "quit", "cancel"},
	})
	if errors.Is(err, terminalexperience.ErrInteractionCancelled) || errors.Is(err, context.Canceled) {
		return "", true, nil
	}
	if errors.Is(err, terminalexperience.ErrAutomationInteraction) {
		return "", false, errExportEnvRequiresInteractive
	}
	if err != nil {
		return "", false, err
	}
	return answer.Value, false, nil
}

func (adapter *terminalExportEnvAdapter) Outro(message string) {
	_ = adapter.run.Present(terminalExportEnvDocument(adapter.session, message, terminalexperience.VisualRoleMuted))
}

func (adapter *terminalExportEnvAdapter) Print(value string) {
	_ = adapter.run.Present(terminalExportEnvDocument(adapter.session, value, terminalexperience.VisualRolePlain))
}

func (adapter *terminalExportEnvAdapter) Cancel(message string) {
	_ = adapter.run.Present(terminalExportEnvDocument(adapter.session, message, terminalexperience.VisualRoleWarning))
}

func exportEnvInteractionOptions(choices []EnvironmentChoice) []terminalexperience.InteractionOption {
	options := make([]terminalexperience.InteractionOption, 0, len(choices))
	for _, choice := range choices {
		options = append(options, terminalexperience.InteractionOption{Label: choice.Label, Value: choice.Value})
	}
	return options
}

func terminalExportEnvDocument(session terminalexperience.Session, text string, role terminalexperience.VisualRole) terminalexperience.PresentationDocument {
	if session.Kind != terminalexperience.RichInteractive {
		role = terminalexperience.VisualRolePlain
	}
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: role, Text: text}}}
}
