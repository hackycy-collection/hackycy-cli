package main

import (
	"context"
	"fmt"
	"os"

	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/updater"
	upgradecommand "github.com/hackycy/hackycy-cli/pkg/cmd/upgrade"
)

func runHiddenUpgrade(arguments []string) (bool, error) {
	if updater.FindInternalMarker(arguments) < 0 {
		return false, nil
	}
	err := updater.RunInternalUpdater(context.Background(), arguments, updater.ReplacementOptions{})
	return true, err
}

func consumeUpgradeStartup(arguments []string, experience *terminalexperience.Runtime) error {
	return consumeUpgradeStartupWith(arguments, experience, os.Executable, updater.ConsumeState)
}

func consumeUpgradeStartupWith(arguments []string, experience *terminalexperience.Runtime, executable func() (string, error), consumeState func(string) (*updater.UpdateTransaction, error)) error {
	if updater.FindInternalMarker(arguments) >= 0 || hasVersionFlag(arguments) {
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
	if state.Status == updater.StatusPending {
		_, _ = fmt.Fprintln(experience.DiagnosticWriter(), upgradecommand.StateMessage(*state))
		return fmt.Errorf("update is still pending")
	}
	return upgradecommand.PresentResult(context.Background(), experience, updater.UpgradeResult{PreviousState: state}, nil)
}

func hasVersionFlag(arguments []string) bool {
	for _, argument := range arguments {
		if argument == "--version" || argument == "-V" {
			return true
		}
	}
	return false
}
