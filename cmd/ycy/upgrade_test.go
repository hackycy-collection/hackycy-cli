package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hackycy/hackycy-cli/internal/commands/upgrade"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
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
	output, diagnostics := &bytes.Buffer{}, &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Session:     terminalexperience.Session{Kind: terminalexperience.Automation},
		Output:      output,
		Diagnostics: diagnostics,
	})
	if err := consumeUpgradeStartup([]string{"--version"}, experience); err != nil {
		t.Fatal(err)
	}
	if output.Len() != 0 || diagnostics.Len() != 0 {
		t.Fatalf("version startup output = %q / %q", output.String(), diagnostics.String())
	}
}

func TestConsumeUpgradeStartupReportsCompletedAndBlocksPending(t *testing.T) {
	testCases := []struct {
		name       string
		status     upgrade.UpdateStatus
		wantOutput string
		wantError  string
	}{
		{name: "completed", status: upgrade.StatusSucceeded, wantOutput: "Updated ycy to v1.0.1.\n"},
		{name: "pending", status: upgrade.StatusPending, wantError: "An update is being applied. Retry in a moment.\n"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			state := testStateForCommand(t)
			state.Status = testCase.status
			output, diagnostics := &bytes.Buffer{}, &bytes.Buffer{}
			experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
				Session:     terminalexperience.Session{Kind: terminalexperience.Automation},
				Output:      output,
				Diagnostics: diagnostics,
			})
			err := consumeUpgradeStartupWith(
				[]string{"config", "cm", "list"},
				experience,
				func() (string, error) { return state.TargetPath, nil },
				func(target string) (*upgrade.UpdateTransaction, error) {
					if target != state.TargetPath {
						t.Fatalf("target = %q, want %q", target, state.TargetPath)
					}
					return &state, nil
				},
			)
			if testCase.status == upgrade.StatusPending {
				if err == nil || err.Error() != "update is still pending" {
					t.Fatalf("pending error = %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if got := output.String(); got != testCase.wantOutput {
				t.Fatalf("stdout = %q, want %q", got, testCase.wantOutput)
			}
			if got := diagnostics.String(); got != testCase.wantError {
				t.Fatalf("stderr = %q, want %q", got, testCase.wantError)
			}
		})
	}
}

func TestTerminalUpgradePresentationPreservesPlainAndAutomationResults(t *testing.T) {
	completed := testStateForCommand(t)
	completed.Status = upgrade.StatusSucceeded
	testCases := []struct {
		name           string
		result         upgrade.UpgradeResult
		err            error
		wantOutput     string
		wantDiagnostic string
	}{
		{
			name:       "completed state and current version",
			result:     upgrade.UpgradeResult{PreviousState: &completed, AlreadyCurrent: true, CurrentVersion: "1.0.1"},
			wantOutput: "Updated ycy to v1.0.1.\nCurrent version v1.0.1 is the latest.\nNo update needed.\n",
		},
		{
			name:       "scheduled",
			result:     upgrade.UpgradeResult{Scheduled: true, ScheduledVersion: "2.0.0"},
			wantOutput: "Update to v2.0.0 has been scheduled and will finish after ycy exits.\n",
		},
		{
			name:           "classified abort is redacted and mixed",
			result:         upgrade.UpgradeResult{Aborted: true},
			err:            &upgrade.ExitCodeError{Code: 0, Err: errors.New("download failed Bearer upgrade-secret")},
			wantOutput:     "Update aborted.\n",
			wantDiagnostic: "error: download failed Bearer [REDACTED]\n",
		},
	}

	for _, testCase := range testCases {
		for _, session := range []terminalexperience.Session{
			{Kind: terminalexperience.PlainInteractive},
			{Kind: terminalexperience.Automation},
		} {
			t.Run(testCase.name, func(t *testing.T) {
				output, diagnostics := &bytes.Buffer{}, &bytes.Buffer{}
				experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
					Session:     session,
					Output:      output,
					Diagnostics: diagnostics,
				})
				err := presentUpgradeResult(context.Background(), experience, testCase.result, testCase.err)
				if testCase.err == nil && err != nil {
					t.Fatalf("Present() error = %v", err)
				}
				if testCase.err != nil && err == nil {
					t.Fatal("classified abort did not preserve its error")
				}
				if got := output.String(); got != testCase.wantOutput {
					t.Fatalf("%v stdout = %q, want %q", session.Kind, got, testCase.wantOutput)
				}
				if got := diagnostics.String(); got != testCase.wantDiagnostic {
					t.Fatalf("%v stderr = %q, want %q", session.Kind, got, testCase.wantDiagnostic)
				}
				if terminaltest.ContainsTerminalControl(output.Bytes()) || terminaltest.ContainsTerminalControl(diagnostics.Bytes()) {
					t.Fatalf("%v Upgrade streams contain terminal control: stdout = %q stderr = %q", session.Kind, output.String(), diagnostics.String())
				}
			})
		}
	}
}

func TestTerminalUpgradePresentationUsesRichSemanticRoles(t *testing.T) {
	for _, testCase := range []struct {
		name string
		role terminalexperience.VisualRole
	}{
		{name: "completed", role: terminalexperience.VisualRoleSuccess},
		{name: "aborted", role: terminalexperience.VisualRoleWarning},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			for _, session := range []terminalexperience.Session{
				{Kind: terminalexperience.RichInteractive, Color: true},
				{Kind: terminalexperience.RichInteractive},
			} {
				document := terminalUpgradeDocument(session, "Update result", testCase.role, true)
				if !document.ClearBefore || len(document.Blocks) != 2 || document.Blocks[0].Role != terminalexperience.VisualRoleTitle || document.Blocks[1].Role != testCase.role {
					t.Fatalf("Rich document = %#v", document)
				}
				var output bytes.Buffer
				experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{Session: session, Output: &output})
				run := experience.Open(context.Background())
				if err := run.Present(document); err != nil {
					t.Fatalf("Present() error = %v", err)
				}
				if err := run.Close(); err != nil {
					t.Fatalf("Close() error = %v", err)
				}
				const clear = "\x1b[2J\x1b[H"
				if !strings.HasPrefix(output.String(), clear) {
					t.Fatalf("Rich output omitted title clear: %q", output.String())
				}
				if !session.Color && strings.Contains(output.String()[len(clear):], "\x1b[") {
					t.Fatalf("NO_COLOR Rich output contains style bytes: %q", output.String())
				}
			}
		})
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
	handler := newUpgradeHandler(terminalexperience.NewExperience(terminalexperience.ExperienceOptions{}), "1.0.0")
	if handler == nil {
		t.Fatal("upgrade handler is nil")
	}
}
