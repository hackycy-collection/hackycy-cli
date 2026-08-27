package main

import (
	"context"
	"errors"

	configcm "github.com/hackycy/hackycy-cli/internal/commands/config/cm"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
)

var errConfigCMAddRequiresInteractive = errors.New("config cm add requires an interactive terminal")

type terminalCMAddAdapter struct {
	run     terminalexperience.ExperienceRun
	session terminalexperience.Session
}

func newTerminalCMAddAdapter(run terminalexperience.ExperienceRun, session terminalexperience.Session) *terminalCMAddAdapter {
	return &terminalCMAddAdapter{run: run, session: session}
}

func (adapter *terminalCMAddAdapter) Text(question configcm.AddTextPrompt) (string, bool, error) {
	return adapter.ask(terminalexperience.InteractionRequest{
		Kind:        terminalexperience.InteractionText,
		Message:     question.Message,
		Placeholder: question.Placeholder,
		Validate: func(answer terminalexperience.InteractionAnswer) error {
			return question.Validate(answer.Value)
		},
	})
}

func (adapter *terminalCMAddAdapter) Password(question configcm.AddTextPrompt) (string, bool, error) {
	return adapter.ask(terminalexperience.InteractionRequest{
		Kind:    terminalexperience.InteractionSecret,
		Message: question.Message,
		Validate: func(answer terminalexperience.InteractionAnswer) error {
			return question.Validate(answer.Value)
		},
	})
}

func (adapter *terminalCMAddAdapter) Cancel(message string) {
	_ = adapter.run.Present(terminalCMAddDocument(adapter.session, message, true))
}

func (adapter *terminalCMAddAdapter) Success(message string) {
	_ = adapter.run.Present(terminalCMAddDocument(adapter.session, message, false))
}

func (adapter *terminalCMAddAdapter) ask(request terminalexperience.InteractionRequest) (string, bool, error) {
	answer, err := adapter.run.Ask(request)
	if errors.Is(err, terminalexperience.ErrInteractionCancelled) || errors.Is(err, context.Canceled) {
		return "", true, nil
	}
	if errors.Is(err, terminalexperience.ErrAutomationInteraction) {
		return "", false, errConfigCMAddRequiresInteractive
	}
	if err != nil {
		return "", false, err
	}
	return answer.Value, false, nil
}

func terminalCMAddDocument(session terminalexperience.Session, message string, cancelled bool) terminalexperience.PresentationDocument {
	role := terminalexperience.VisualRolePlain
	if session.Kind == terminalexperience.RichInteractive {
		role = terminalexperience.VisualRoleSuccess
		if cancelled {
			role = terminalexperience.VisualRoleWarning
		}
	}
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: role, Text: message}}}
}
