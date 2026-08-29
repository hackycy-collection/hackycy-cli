package main

import (
	"bytes"
	"context"
	"os"
	"reflect"
	"strings"
	"testing"

	configfork "github.com/hackycy/hackycy-cli/internal/commands/config/fork"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestTerminalForkListPresentationPreservesPlainAndAutomationResults(t *testing.T) {
	result := configfork.Result{Instances: []configfork.Instance{{
		Name:         "work",
		Host:         "gitlab.example",
		Scheme:       "https",
		Type:         "gitlab",
		TokenPreview: "MDEy***",
	}}}
	const want = "NAME  TYPE    SCHEME  HOST            TOKEN\nwork  gitlab  https   gitlab.example  MDEy***\n1 instance configured\n"

	for _, session := range []terminalexperience.Session{
		{Kind: terminalexperience.PlainInteractive},
		{Kind: terminalexperience.Automation},
	} {
		var output bytes.Buffer
		experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{Session: session, Output: &output})
		run := experience.Open(context.Background())
		if err := run.Present(terminalForkListDocument(session, result)); err != nil {
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

func TestTerminalForkListPresentationUsesRichSemanticRoles(t *testing.T) {
	result := configfork.Result{Instances: []configfork.Instance{{
		Name:         "work",
		Host:         "gitlab.example",
		Scheme:       "https",
		Type:         "gitlab",
		TokenPreview: "MDEy***",
	}}}

	for _, testCase := range []struct {
		name    string
		session terminalexperience.Session
	}{
		{name: "color", session: terminalexperience.Session{Kind: terminalexperience.RichInteractive, Color: true}},
		{name: "no color", session: terminalexperience.Session{Kind: terminalexperience.RichInteractive}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			document := terminalForkListDocument(testCase.session, result)
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
			for _, expected := range []string{"Fork provider instances", "NAME  TYPE  SCHEME  HOST  TOKEN", "work", "gitlab.example", "MDEy***", "1 instance configured"} {
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

func environmentWith(overrides map[string]string) []string {
	environment := make([]string, 0, len(os.Environ())+len(overrides))
	for _, entry := range os.Environ() {
		key, _, found := strings.Cut(entry, "=")
		if !found {
			continue
		}
		if _, replaced := overrides[key]; !replaced {
			environment = append(environment, entry)
		}
	}
	for key, value := range overrides {
		environment = append(environment, key+"="+value)
	}
	return environment
}
