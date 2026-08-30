package upgrade

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
	"github.com/hackycy/hackycy-cli/internal/updater"
)

func TestTerminalUpgradePresentationPreservesPlainAndAutomationResults(t *testing.T) {
	completed := updater.UpdateTransaction{ExpectedVersion: "1.0.1", Status: updater.StatusSucceeded}
	testCases := []struct {
		name           string
		result         updater.UpgradeResult
		err            error
		wantOutput     string
		wantDiagnostic string
	}{
		{
			name:       "completed state and current version",
			result:     updater.UpgradeResult{PreviousState: &completed, AlreadyCurrent: true, CurrentVersion: "1.0.1"},
			wantOutput: "Updated ycy to v1.0.1.\nCurrent version v1.0.1 is the latest.\nNo update needed.\n",
		},
		{
			name:       "scheduled",
			result:     updater.UpgradeResult{Scheduled: true, ScheduledVersion: "2.0.0"},
			wantOutput: "Update to v2.0.0 has been scheduled and will finish after ycy exits.\n",
		},
		{
			name:           "classified abort is redacted and mixed",
			result:         updater.UpgradeResult{Aborted: true},
			err:            &updater.ExitCodeError{Code: 0, Err: errors.New("download failed Bearer upgrade-secret")},
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
				err := PresentResult(context.Background(), experience, testCase.result, testCase.err)
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
