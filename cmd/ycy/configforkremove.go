package main

import (
	"context"
	"errors"

	configfork "github.com/hackycy/hackycy-cli/internal/commands/config/fork"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
)

var errConfigForkRemoveRequiresInteractive = errors.New("config fork remove requires an interactive terminal")

type terminalForkRemoveAdapter struct {
	run     terminalexperience.ExperienceRun
	session terminalexperience.Session
}

func newTerminalForkRemoveAdapter(run terminalexperience.ExperienceRun, session terminalexperience.Session) *terminalForkRemoveAdapter {
	return &terminalForkRemoveAdapter{run: run, session: session}
}

func (adapter *terminalForkRemoveAdapter) Select(question configfork.SelectPrompt) (string, bool, error) {
	request := terminalexperience.InteractionRequest{
		Kind:    terminalexperience.InteractionSelect,
		Message: question.Message,
		Options: interactionOptions(question.Choices),
	}
	if len(question.Choices) > 0 {
		request.HasDefault = true
		request.Default = terminalexperience.InteractionAnswer{Value: question.Choices[0].Value}
	}
	answer, cancelled, err := adapter.ask(request)
	return answer.Value, cancelled, err
}

func (adapter *terminalForkRemoveAdapter) Confirm(question configfork.ConfirmPrompt) (bool, bool, error) {
	answer, cancelled, err := adapter.ask(terminalexperience.InteractionRequest{
		Kind:       terminalexperience.InteractionConfirm,
		Message:    question.Message,
		HasDefault: true,
		Default:    terminalexperience.InteractionAnswer{Confirmed: false},
	})
	return answer.Confirmed, cancelled, err
}

func (adapter *terminalForkRemoveAdapter) Info(message string) {
	_ = adapter.run.Present(terminalForkRemoveDocument(adapter.session, message, terminalexperience.VisualRoleMuted))
}

func (adapter *terminalForkRemoveAdapter) Outcome(message string) {
	role := terminalexperience.VisualRolePlain
	if adapter.session.Kind == terminalexperience.RichInteractive {
		role = terminalexperience.VisualRoleSuccess
		if message == "Cancelled" {
			role = terminalexperience.VisualRoleWarning
		}
	}
	_ = adapter.run.Present(terminalForkRemoveDocument(adapter.session, message, role))
}

func (adapter *terminalForkRemoveAdapter) ask(request terminalexperience.InteractionRequest) (terminalexperience.InteractionAnswer, bool, error) {
	answer, err := adapter.run.Ask(request)
	if errors.Is(err, terminalexperience.ErrInteractionCancelled) || errors.Is(err, context.Canceled) {
		return terminalexperience.InteractionAnswer{}, true, nil
	}
	if errors.Is(err, terminalexperience.ErrAutomationInteraction) {
		return terminalexperience.InteractionAnswer{}, false, errConfigForkRemoveRequiresInteractive
	}
	if err != nil {
		return terminalexperience.InteractionAnswer{}, false, err
	}
	return answer, false, nil
}

func terminalForkRemoveDocument(session terminalexperience.Session, message string, role terminalexperience.VisualRole) terminalexperience.PresentationDocument {
	if session.Kind != terminalexperience.RichInteractive {
		role = terminalexperience.VisualRolePlain
	}
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: role, Text: message}}}
}
