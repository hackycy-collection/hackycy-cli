package add

import (
	"context"
	"errors"

	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
)

type terminalCMAddAdapter struct {
	run terminalexperience.ExperienceRun
}

func newTerminalCMAddAdapter(run terminalexperience.ExperienceRun) *terminalCMAddAdapter {
	return &terminalCMAddAdapter{run: run}
}

func (adapter *terminalCMAddAdapter) Text(question AddTextPrompt) (string, bool, error) {
	return adapter.ask(terminalexperience.InteractionRequest{Kind: terminalexperience.InteractionText, Message: question.Message, Placeholder: question.Placeholder, Validate: func(answer terminalexperience.InteractionAnswer) error { return question.Validate(answer.Value) }})
}

func (adapter *terminalCMAddAdapter) Password(question AddTextPrompt) (string, bool, error) {
	return adapter.ask(terminalexperience.InteractionRequest{Kind: terminalexperience.InteractionSecret, Message: question.Message, Validate: func(answer terminalexperience.InteractionAnswer) error { return question.Validate(answer.Value) }})
}

func (adapter *terminalCMAddAdapter) Cancel(message string) {
	_ = adapter.run.Result(terminalCMAddDocument(message, true))
}
func (adapter *terminalCMAddAdapter) Success(message string) {
	_ = adapter.run.Result(terminalCMAddDocument(message, false))
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

func terminalCMAddDocument(message string, cancelled bool) terminalexperience.PresentationDocument {
	role := terminalexperience.VisualRoleSuccess
	if cancelled {
		role = terminalexperience.VisualRoleWarning
	}
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: role, Text: message}}}
}
