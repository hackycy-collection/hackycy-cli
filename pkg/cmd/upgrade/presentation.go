package upgrade

import (
	"context"
	"errors"
	"fmt"

	"github.com/hackycy/hackycy-cli/internal/logging"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/updater"
)

// PresentResult preserves Upgrade's terminal result and redaction behavior.
func PresentResult(ctx context.Context, experience *terminalexperience.Runtime, result updater.UpgradeResult, resultErr error) error {
	run := experience.Open(ctx)
	defer run.Close()
	cleared := false
	present := func(message string, role terminalexperience.VisualRole) error {
		if err := run.Present(terminalUpgradeDocument(experience.Session(), message, role, !cleared)); err != nil {
			return err
		}
		cleared = true
		return nil
	}

	if result.PreviousState != nil {
		if err := present(StateMessage(*result.PreviousState), terminalUpgradeStateRole(*result.PreviousState)); err != nil {
			return err
		}
	}
	if resultErr != nil {
		var exit *updater.ExitCodeError
		if result.Aborted && errors.As(resultErr, &exit) {
			_, _ = fmt.Fprintln(experience.DiagnosticWriter(), "error: "+logging.Redact(resultErr.Error()))
			if err := present("Update aborted.", terminalexperience.VisualRoleWarning); err != nil {
				return err
			}
		}
		return resultErr
	}
	if result.AlreadyCurrent {
		return present(fmt.Sprintf("Current version v%s is the latest.\nNo update needed.", result.CurrentVersion), terminalexperience.VisualRoleSuccess)
	}
	if result.Scheduled {
		return present(fmt.Sprintf("Update to v%s has been scheduled and will finish after ycy exits.", result.ScheduledVersion), terminalexperience.VisualRoleSuccess)
	}
	return nil
}

func terminalUpgradeDocument(session terminalexperience.Session, message string, role terminalexperience.VisualRole, clear bool) terminalexperience.PresentationDocument {
	if session.Kind != terminalexperience.RichInteractive {
		return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{
			Role: terminalexperience.VisualRolePlain,
			Text: message,
		}}}
	}
	blocks := make([]terminalexperience.PresentationBlock, 0, 2)
	if clear {
		blocks = append(blocks, terminalexperience.PresentationBlock{
			Role: terminalexperience.VisualRoleTitle,
			Text: "HACKYCY CLI",
		})
	}
	blocks = append(blocks, terminalexperience.PresentationBlock{Role: role, Text: message})
	return terminalexperience.PresentationDocument{ClearBefore: clear, Blocks: blocks}
}

// StateMessage formats a persisted result for startup presentation.
func StateMessage(state updater.UpdateTransaction) string {
	return logging.Redact(updater.FormatStateResult(state))
}

func terminalUpgradeStateRole(state updater.UpdateTransaction) terminalexperience.VisualRole {
	switch state.Status {
	case updater.StatusSucceeded:
		return terminalexperience.VisualRoleSuccess
	case updater.StatusSucceededCleanupWarn, updater.StatusFailed, updater.StatusPending:
		return terminalexperience.VisualRoleWarning
	default:
		return terminalexperience.VisualRolePlain
	}
}
