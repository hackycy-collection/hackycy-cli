package main

import (
	"bytes"
	"context"
	"testing"

	configcm "github.com/hackycy/hackycy-cli/internal/commands/config/cm"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestTerminalCMTestPresentationPreservesPlainAndAutomationResults(t *testing.T) {
	tests := []struct {
		name   string
		result configcm.TestResult
		want   string
	}{
		{name: "success", result: configcm.TestResult{Content: "ok"}, want: "Response: ok\nDone\n"},
		{name: "failure", result: configcm.TestResult{Diagnostic: &configcm.TestDiagnostic{Provider: "work", BaseURL: "https://provider.test/v1", Model: "provider-model"}}, want: "Provider: work\nBase URL: https://provider.test/v1\nModel: provider-model\n"},
	}
	for _, testCase := range tests {
		for _, session := range []terminalexperience.Session{
			{Kind: terminalexperience.PlainInteractive},
			{Kind: terminalexperience.Automation},
		} {
			var output bytes.Buffer
			experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{Session: session, Output: &output})
			run := experience.Open(context.Background())
			if err := run.Present(terminalCMTestDocument(session, testCase.result)); err != nil {
				t.Fatalf("%s Present() error = %v", testCase.name, err)
			}
			if err := run.Close(); err != nil {
				t.Fatalf("%s Close() error = %v", testCase.name, err)
			}
			if got := output.String(); got != testCase.want {
				t.Fatalf("%s %v output = %q, want %q", testCase.name, session.Kind, got, testCase.want)
			}
			if terminaltest.ContainsTerminalControl(output.Bytes()) {
				t.Fatalf("%s %v output contains terminal control: %q", testCase.name, session.Kind, output.String())
			}
		}
	}
}

func TestTerminalCMTestPresentationUsesRichSemanticRoles(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		result configcm.TestResult
		roles  []terminalexperience.VisualRole
	}{
		{name: "success", result: configcm.TestResult{Content: "ok"}, roles: []terminalexperience.VisualRole{terminalexperience.VisualRoleTitle, terminalexperience.VisualRolePlain, terminalexperience.VisualRoleSuccess}},
		{name: "failure", result: configcm.TestResult{Diagnostic: &configcm.TestDiagnostic{Provider: "work", BaseURL: "https://provider.test/v1", Model: "provider-model"}}, roles: []terminalexperience.VisualRole{terminalexperience.VisualRoleTitle, terminalexperience.VisualRoleWarning, terminalexperience.VisualRoleMuted}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			for _, session := range []terminalexperience.Session{
				{Kind: terminalexperience.RichInteractive, Color: true},
				{Kind: terminalexperience.RichInteractive},
			} {
				document := terminalCMTestDocument(session, testCase.result)
				if len(document.Blocks) != len(testCase.roles) {
					t.Fatalf("Rich blocks = %#v", document.Blocks)
				}
				for index, role := range testCase.roles {
					if document.Blocks[index].Role != role {
						t.Fatalf("Rich block %d role = %v, want %v", index, document.Blocks[index].Role, role)
					}
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
				if terminaltest.ContainsTerminalControl(output.Bytes()) {
					t.Fatalf("non-terminal writer output contains terminal control: %q", output.String())
				}
			}
		})
	}
}
