package main

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	configcm "github.com/hackycy/hackycy-cli/internal/commands/config/cm"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestTerminalCMListPresentationPreservesPlainAndAutomationResults(t *testing.T) {
	result := configcm.Result{Profiles: []configcm.Profile{
		{Name: "work", Model: "gpt-4.1-mini", BaseURL: "https://work.example/v1"},
		{Name: "personal", Model: "deepseek-chat", BaseURL: "https://personal.example/v1", Default: true},
	}}
	const want = "  work gpt-4.1-mini https://work.example/v1\n* personal deepseek-chat https://personal.example/v1\n"

	for _, session := range []terminalexperience.Session{
		{Kind: terminalexperience.PlainInteractive},
		{Kind: terminalexperience.Automation},
	} {
		var output bytes.Buffer
		experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{Session: session, Output: &output})
		run := experience.Open(context.Background())
		if err := run.Present(terminalCMListDocument(session, result)); err != nil {
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

func TestTerminalCMListPresentationUsesRichSemanticRoles(t *testing.T) {
	result := configcm.Result{Profiles: []configcm.Profile{
		{Name: "work", Model: "gpt-4.1-mini", BaseURL: "https://work.example/v1"},
		{Name: "personal", Model: "deepseek-chat", BaseURL: "https://personal.example/v1", Default: true},
	}}

	for _, testCase := range []struct {
		name    string
		session terminalexperience.Session
	}{
		{name: "color", session: terminalexperience.Session{Kind: terminalexperience.RichInteractive, Color: true}},
		{name: "no color", session: terminalexperience.Session{Kind: terminalexperience.RichInteractive}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			document := terminalCMListDocument(testCase.session, result)
			if got, want := []terminalexperience.VisualRole{
				document.Blocks[0].Role,
				document.Blocks[1].Role,
				document.Blocks[2].Role,
				document.Blocks[3].Role,
			}, []terminalexperience.VisualRole{
				terminalexperience.VisualRoleTitle,
				terminalexperience.VisualRoleMuted,
				terminalexperience.VisualRolePlain,
				terminalexperience.VisualRoleSuccess,
			}; !reflect.DeepEqual(got, want) {
				t.Fatalf("Rich roles = %#v, want %#v", got, want)
			}
			var output bytes.Buffer
			experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{Session: testCase.session, Output: &output})
			run := experience.Open(context.Background())
			if err := run.Present(document); err != nil {
				t.Fatalf("Present() error = %v", err)
			}
			if err := run.Close(); err != nil {
				t.Fatalf("Close() error = %v", err)
			}
			for _, expected := range []string{"Commit message profiles", "PROFILE  MODEL  BASE URL", "work", "personal", "deepseek-chat"} {
				if !strings.Contains(output.String(), expected) {
					t.Fatalf("Rich result = %q, missing %q", output.String(), expected)
				}
			}
			if terminaltest.ContainsTerminalControl(output.Bytes()) {
				t.Fatalf("non-terminal writer output contains terminal control: %q", output.String())
			}
		})
	}
}
