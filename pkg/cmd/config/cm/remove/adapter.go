package remove

import (
	"context"
	"errors"

	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
)

var errConfigCMRemoveRequiresInteractive = errors.New("config cm remove requires an interactive terminal")

type terminalCMRemoveAdapter struct {
	run     terminalexperience.ExperienceRun
	session terminalexperience.Session
}

func newTerminalCMRemoveAdapter(run terminalexperience.ExperienceRun, session terminalexperience.Session) *terminalCMRemoveAdapter {
	return &terminalCMRemoveAdapter{run: run, session: session}
}

func (adapter *terminalCMRemoveAdapter) Confirm(question RemoveConfirmPrompt) (bool, bool, error) {
	answer, err := adapter.run.Ask(terminalexperience.InteractionRequest{Kind: terminalexperience.InteractionConfirm, Message: question.Message, HasDefault: true, Default: terminalexperience.InteractionAnswer{Confirmed: false}})
	if errors.Is(err, terminalexperience.ErrInteractionCancelled) || errors.Is(err, context.Canceled) {
		return false, true, nil
	}
	if errors.Is(err, terminalexperience.ErrAutomationInteraction) {
		return false, false, errConfigCMRemoveRequiresInteractive
	}
	if err != nil {
		return false, false, err
	}
	return answer.Confirmed, false, nil
}

func (adapter *terminalCMRemoveAdapter) Cancel(message string) {
	_ = adapter.run.Present(terminalCMRemoveDocument(adapter.session, message, true))
}
func (adapter *terminalCMRemoveAdapter) Success(message string) {
	_ = adapter.run.Present(terminalCMRemoveDocument(adapter.session, message, false))
}

func terminalCMRemoveDocument(session terminalexperience.Session, message string, cancelled bool) terminalexperience.PresentationDocument {
	role := terminalexperience.VisualRolePlain
	if session.Kind == terminalexperience.RichInteractive {
		role = terminalexperience.VisualRoleSuccess
		if cancelled {
			role = terminalexperience.VisualRoleWarning
		}
	}
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: role, Text: message}}}
}
