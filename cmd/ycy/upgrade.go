package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/hackycy/hackycy-cli/internal/cliapp"
	"github.com/hackycy/hackycy-cli/internal/commands/upgrade"
)

func newUpgradeHandler(output, errorOutput io.Writer, currentVersion string) cliapp.UpgradeHandler {
	return func(ctx context.Context) error {
		_, err := upgrade.RunUpgrade(ctx, upgrade.UpgradeOptions{
			Resolver:    upgrade.ReleaseResolverOptions{CurrentVersion: currentVersion},
			Output:      output,
			ErrorOutput: errorOutput,
		})
		return err
	}
}

func runHiddenUpgrade(arguments []string) (bool, error) {
	if upgrade.FindInternalMarker(arguments) < 0 {
		return false, nil
	}
	err := upgrade.RunInternalUpdater(context.Background(), arguments, upgrade.ReplacementOptions{})
	return true, err
}

func consumeUpgradeStartup(arguments []string, output, errorOutput io.Writer) error {
	if upgrade.FindInternalMarker(arguments) >= 0 || hasVersionFlag(arguments) {
		return nil
	}
	target, err := os.Executable()
	if err != nil {
		return err
	}
	state, err := upgrade.ConsumeState(target)
	if err != nil {
		return err
	}
	if state == nil {
		return nil
	}
	if state.Status == upgrade.StatusPending {
		_, _ = fmt.Fprintln(errorOutput, upgrade.FormatStateResult(*state))
		return fmt.Errorf("update is still pending")
	}
	_, _ = fmt.Fprintln(output, upgrade.FormatStateResult(*state))
	return nil
}

func hasVersionFlag(arguments []string) bool {
	for _, argument := range arguments {
		if argument == "--version" || argument == "-V" {
			return true
		}
	}
	return false
}
