package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/hackycy/hackycy-cli/internal/cliapp"
	"github.com/hackycy/hackycy-cli/internal/commands/upgrade"
	"github.com/hackycy/hackycy-cli/internal/logging"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
)

func newUpgradeHandler(experience *terminalexperience.Runtime, currentVersion string) cliapp.UpgradeHandler {
	return func(ctx context.Context) error {
		result, err := upgrade.RunUpgrade(ctx, upgrade.UpgradeOptions{
			Resolver: upgrade.ReleaseResolverOptions{CurrentVersion: currentVersion},
		})
		return presentUpgradeResult(ctx, experience, result, err)
	}
}

func runHiddenUpgrade(arguments []string) (bool, error) {
	if upgrade.FindInternalMarker(arguments) < 0 {
		return false, nil
	}
	err := upgrade.RunInternalUpdater(context.Background(), arguments, upgrade.ReplacementOptions{})
	return true, err
}

func consumeUpgradeStartup(arguments []string, experience *terminalexperience.Runtime) error {
	return consumeUpgradeStartupWith(arguments, experience, os.Executable, upgrade.ConsumeState)
}

func consumeUpgradeStartupWith(arguments []string, experience *terminalexperience.Runtime, executable func() (string, error), consumeState func(string) (*upgrade.UpdateTransaction, error)) error {
	if upgrade.FindInternalMarker(arguments) >= 0 || hasVersionFlag(arguments) {
		return nil
	}
	target, err := executable()
	if err != nil {
		return err
	}
	state, err := consumeState(target)
	if err != nil {
		return err
	}
	if state == nil {
		return nil
	}
	if state.Status == upgrade.StatusPending {
		_, _ = fmt.Fprintln(experience.DiagnosticWriter(), terminalUpgradeStateMessage(*state))
		return fmt.Errorf("update is still pending")
	}
	return presentUpgradeResult(context.Background(), experience, upgrade.UpgradeResult{PreviousState: state}, nil)
}

func presentUpgradeResult(ctx context.Context, experience *terminalexperience.Runtime, result upgrade.UpgradeResult, resultErr error) error {
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
		if err := present(terminalUpgradeStateMessage(*result.PreviousState), terminalUpgradeStateRole(*result.PreviousState)); err != nil {
			return err
		}
	}
	if resultErr != nil {
		var exit *upgrade.ExitCodeError
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

func terminalUpgradeStateMessage(state upgrade.UpdateTransaction) string {
	return logging.Redact(upgrade.FormatStateResult(state))
}

func terminalUpgradeStateRole(state upgrade.UpdateTransaction) terminalexperience.VisualRole {
	switch state.Status {
	case upgrade.StatusSucceeded:
		return terminalexperience.VisualRoleSuccess
	case upgrade.StatusSucceededCleanupWarn, upgrade.StatusFailed, upgrade.StatusPending:
		return terminalexperience.VisualRoleWarning
	default:
		return terminalexperience.VisualRolePlain
	}
}

func hasVersionFlag(arguments []string) bool {
	for _, argument := range arguments {
		if argument == "--version" || argument == "-V" {
			return true
		}
	}
	return false
}
