package main

import (
	"bytes"
	"context"
	"testing"

	configcm "github.com/hackycy/hackycy-cli/internal/commands/config/cm"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestTerminalCMSetPresentationPreservesPlainAndAutomationResults(t *testing.T) {
	result := configcm.SetResult{Profile: "work"}
	const want = "Profile work updated\n"

	for _, session := range []terminalexperience.Session{
		{Kind: terminalexperience.PlainInteractive},
		{Kind: terminalexperience.Automation},
	} {
		var output bytes.Buffer
		experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{Session: session, Output: &output})
		run := experience.Open(context.Background())
		if err := run.Present(terminalCMSetDocument(session, result)); err != nil {
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

func TestTerminalCMSetPresentationUsesARichSuccessRole(t *testing.T) {
	for _, session := range []terminalexperience.Session{
		{Kind: terminalexperience.RichInteractive, Color: true},
		{Kind: terminalexperience.RichInteractive},
	} {
		document := terminalCMSetDocument(session, configcm.SetResult{Profile: "work"})
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
		if got, want := output.String(), "Profile work updated\n"; got != want {
			t.Fatalf("Rich output = %q, want %q", got, want)
		}
		if terminaltest.ContainsTerminalControl(output.Bytes()) {
			t.Fatalf("non-terminal writer output contains terminal control: %q", output.String())
		}
	}
}
