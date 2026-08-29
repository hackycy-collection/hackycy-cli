package main

import (
	"context"
	"errors"
	"os"

	"github.com/hackycy/hackycy-cli/internal/commands/exportenv"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	rootcommand "github.com/hackycy/hackycy-cli/pkg/cmd/root"
)

var errExportEnvRequiresInteractive = errors.New("export env requires an interactive terminal")

func newExportEnvHandler(experience *terminalexperience.Runtime) rootcommand.ExportEnvHandler {
	return func(ctx context.Context, input exportenv.Input) (exportenv.Result, error) {
		run := experience.Open(ctx)
		defer run.Close()
		adapter := newTerminalExportEnvAdapter(run, experience.Session())
		module, err := exportenv.New(exportenv.Dependencies{
			WorkingDirectory: os.Getwd,
			Selector:         adapter,
			Reader:           osExportEnvReader{},
			Writer:           osExportEnvWriter{},
			Presenter:        adapter,
		})
		if err != nil {
			return exportenv.Result{}, err
		}
		return module.Run(ctx, input)
	}
}

type terminalExportEnvAdapter struct {
	run     terminalexperience.ExperienceRun
	session terminalexperience.Session
}

func newTerminalExportEnvAdapter(run terminalexperience.ExperienceRun, session terminalexperience.Session) *terminalExportEnvAdapter {
	return &terminalExportEnvAdapter{run: run, session: session}
}

func (adapter *terminalExportEnvAdapter) SelectEnvironment(message string, choices []exportenv.EnvironmentChoice) (string, bool, error) {
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

func exportEnvInteractionOptions(choices []exportenv.EnvironmentChoice) []terminalexperience.InteractionOption {
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

type osExportEnvReader struct{}

func (osExportEnvReader) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

type osExportEnvWriter struct{}

func (osExportEnvWriter) WriteFile(path string, content []byte) error {
	return os.WriteFile(path, content, 0o666)
}
