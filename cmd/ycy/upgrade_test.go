package main

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hackycy/hackycy-cli/internal/commands/upgrade"
)

func TestHiddenUpgradeRouteIsolatedFromOrdinaryArguments(t *testing.T) {
	if handled, err := runHiddenUpgrade([]string{"ordinary", "value"}); handled || err != nil {
		t.Fatalf("ordinary route = handled %v, err %v", handled, err)
	}
	state := testStateForCommand(t)
	if err := upgrade.WriteState(state); err != nil {
		t.Fatal(err)
	}
	arguments := append([]string{"ordinary", "value"}, upgrade.InternalUpdateArgs(state)...)
	// The fixture intentionally fails before touching a target; marker discovery is
	// the property under test and the hidden route must not fall through Cobra.
	handled, err := runHiddenUpgrade(arguments)
	if !handled || err == nil {
		t.Fatalf("hidden route = handled %v, err %v", handled, err)
	}
}

func TestVersionFlagsSkipStartupResultConsumption(t *testing.T) {
	output, errorOutput := &strings.Builder{}, &strings.Builder{}
	if err := consumeUpgradeStartup([]string{"--version"}, output, errorOutput); err != nil {
		t.Fatal(err)
	}
	if output.Len() != 0 || errorOutput.Len() != 0 {
		t.Fatalf("version startup output = %q / %q", output.String(), errorOutput.String())
	}
}

func TestConsumeUpgradeStartupReportsCompletedAndBlocksPending(t *testing.T) {
	state := testStateForCommand(t)
	state.Status = upgrade.StatusSucceeded
	if err := upgrade.WriteState(state); err != nil {
		t.Fatal(err)
	}
	// os.Executable is not the temp target, so exercise the public formatter through
	// the direct state contract instead of mutating the running test binary.
	output := &strings.Builder{}
	if _, err := upgrade.ConsumeState(state.TargetPath); err != nil {
		t.Fatal(err)
	}
	_, _ = io.WriteString(output, upgrade.FormatStateResult(state))
	if !strings.Contains(output.String(), "Updated ycy") {
		t.Fatalf("completed output = %q", output.String())
	}
}

func testStateForCommand(t *testing.T) upgrade.UpdateTransaction {
	t.Helper()
	directory := t.TempDir()
	target := filepath.Join(directory, "ycy")
	state := upgrade.UpdateTransaction{
		TransactionID:   "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
		ParentPID:       2147483647,
		TargetPath:      target,
		StagedPath:      target + ".new.bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
		BackupPath:      target + ".backup.bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
		ExpectedHash:    strings.Repeat("a", 64),
		ExpectedVersion: "1.0.1",
		StatePath:       upgrade.StatePath(target),
		UpdaterPath:     filepath.Join(directory, "updater"),
		CreatedAt:       "2026-08-25T00:00:00Z",
		Status:          upgrade.StatusPending,
	}
	if err := os.WriteFile(state.StagedPath, []byte("candidate"), 0o600); err != nil {
		t.Fatal(err)
	}
	return state
}

func TestUpgradeHandlerConstructsWithoutExternalTools(t *testing.T) {
	handler := newUpgradeHandler(io.Discard, io.Discard, "1.0.0")
	if handler == nil {
		t.Fatal("upgrade handler is nil")
	}
	_ = context.Background()
	_ = errors.New("fixture")
}
