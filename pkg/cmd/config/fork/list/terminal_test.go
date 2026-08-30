package list

import (
	"bytes"
	"context"
	"os"
	"reflect"
	"strings"
	"testing"

	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestTerminalForkListPresentationPreservesPlainAndAutomationResults(t *testing.T) {
	result := Result{Instances: []Instance{{
		Name:         "work",
		Host:         "gitlab.example",
		Scheme:       "https",
		Type:         "gitlab",
		TokenPreview: "MDEy***",
	}}}
	const want = "Fork provider instances\nNAME  TYPE  SCHEME  HOST  TOKEN\nwork  gitlab  https  gitlab.example  MDEy***\n1 instance configured\n"

	for _, session := range []terminalexperience.Capabilities{
		{Interaction: terminalexperience.PlainInteractive},
		{Interaction: terminalexperience.Automation},
	} {
		var output bytes.Buffer
		experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{Capabilities: session, Output: &output})
		run := experience.Open(context.Background())
		if err := run.Result(terminalForkListDocument(result)); err != nil {
			t.Fatalf("Present() error = %v", err)
		}
		if err := run.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
		if got := output.String(); got != want {
			t.Fatalf("%v result = %q, want %q", session.Interaction, got, want)
		}
		if terminaltest.ContainsTerminalControl(output.Bytes()) {
			t.Fatalf("%v result contains terminal control: %q", session.Interaction, output.String())
		}
	}
}

func TestTerminalForkListPresentationUsesRichSemanticRoles(t *testing.T) {
	result := Result{Instances: []Instance{{
		Name:         "work",
		Host:         "gitlab.example",
		Scheme:       "https",
		Type:         "gitlab",
		TokenPreview: "MDEy***",
	}}}

	for _, testCase := range []struct {
		name    string
		session terminalexperience.Capabilities
	}{
		{name: "color", session: terminalexperience.Capabilities{Interaction: terminalexperience.RichInteractive}},
		{name: "no color", session: terminalexperience.Capabilities{Interaction: terminalexperience.RichInteractive}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			document := terminalForkListDocument(result)
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
			experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{Capabilities: testCase.session, Output: &output})
			run := experience.Open(context.Background())
			if err := run.Result(document); err != nil {
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
