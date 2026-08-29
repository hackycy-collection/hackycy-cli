package use

import (
	"bytes"
	"context"
	"testing"

	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestTerminalCMUsePresentationPreservesPlainAndAutomationResults(t *testing.T) {
	result := UseResult{Profile: "work"}
	const want = "Default CM profile set to work\n"

	for _, session := range []terminalexperience.Session{
		{Kind: terminalexperience.PlainInteractive},
		{Kind: terminalexperience.Automation},
	} {
		var output bytes.Buffer
		experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{Session: session, Output: &output})
		run := experience.Open(context.Background())
		if err := run.Present(terminalCMUseDocument(session, result)); err != nil {
			t.Fatalf("Present() error = %v", err)
		}
		if err := run.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
		if got := output.String(); got != want {
			t.Fatalf("%v result = %q, want %q", session.Kind, got, want)
		}
		if terminaltest.ContainsTerminalControl(output.Bytes()) {
			t.Fatalf("%v result contains terminal control: %q", session.Kind, output.String())
		}
	}
}

func TestTerminalCMUsePresentationUsesARichSuccessRole(t *testing.T) {
	result := UseResult{Profile: "work"}
	for _, session := range []terminalexperience.Session{
		{Kind: terminalexperience.RichInteractive, Color: true},
		{Kind: terminalexperience.RichInteractive},
	} {
		document := terminalCMUseDocument(session, result)
		if got, want := document.Blocks[0].Role, terminalexperience.VisualRoleSuccess; got != want {
			t.Fatalf("Rich role = %v, want %v", got, want)
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
		if got, want := output.String(), "Default CM profile set to work\n"; got != want {
			t.Fatalf("Rich output = %q, want %q", got, want)
		}
		if terminaltest.ContainsTerminalControl(output.Bytes()) {
			t.Fatalf("non-terminal writer output contains terminal control: %q", output.String())
		}
	}
}
