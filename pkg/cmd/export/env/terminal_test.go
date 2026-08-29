package env

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestTerminalExportEnvAdapterTranslatesSelectionAndPresentation(t *testing.T) {
	experience := terminaltest.NewRecordingExperience(terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Value: ".env.production"}})
	run := experience.Open(context.Background())
	adapter := newTerminalExportEnvAdapter(run, terminalexperience.Session{Kind: terminalexperience.RichInteractive, Color: true})
	choices := []EnvironmentChoice{
		{Value: ".env", Label: "default"},
		{Value: ".env.production", Label: "production"},
	}

	value, cancelled, err := adapter.SelectEnvironment("Select environment", choices)
	if err != nil || cancelled || value != ".env.production" {
		t.Fatalf("SelectEnvironment() = (%q, %t, %v)", value, cancelled, err)
	}
	adapter.Outro("Exported variables:")
	adapter.Print("{\n  \"VALUE\": \"production\"\n}")
	adapter.Cancel("Cancelled")
	if err := run.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	operations := experience.Run.Operations()
	if len(operations) != 5 || operations[0].Kind != terminaltest.AskOperation || operations[4].Kind != terminaltest.CloseOperation {
		t.Fatalf("operations = %#v", operations)
	}
	request := operations[0].Value.(terminalexperience.InteractionRequest)
	if request.Kind != terminalexperience.InteractionSelect || request.Message != "Select environment" || request.HasDefault || !reflect.DeepEqual(request.Options, []terminalexperience.InteractionOption{
		{Value: ".env", Label: "default"},
		{Value: ".env.production", Label: "production"},
	}) || !reflect.DeepEqual(request.CancelValues, []string{"", "q", "quit", "cancel"}) {
		t.Fatalf("selection request = %#v", request)
	}
	for index, want := range []terminalexperience.PresentationDocument{
		{Blocks: []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleMuted, Text: "Exported variables:"}}},
		{Blocks: []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRolePlain, Text: "{\n  \"VALUE\": \"production\"\n}"}}},
		{Blocks: []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleWarning, Text: "Cancelled"}}},
	} {
		if operations[index+1].Kind != terminaltest.PresentOperation || !reflect.DeepEqual(operations[index+1].Value, want) {
			t.Fatalf("presentation %d = %#v, want %#v", index, operations[index+1], want)
		}
	}
}

func TestTerminalExportEnvAdapterRoutesPlainSelectionValidationAndCancellation(t *testing.T) {
	choices := []EnvironmentChoice{
		{Value: ".env.local", Label: "local"},
		{Value: ".env.production", Label: "production"},
	}
	stdout, diagnostics := &bytes.Buffer{}, &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Session:     terminalexperience.Session{Kind: terminalexperience.PlainInteractive},
		Input:       strings.NewReader("invalid\n2\n"),
		Output:      stdout,
		Diagnostics: diagnostics,
	})
	run := experience.Open(context.Background())
	adapter := newTerminalExportEnvAdapter(run, experience.Session())
	value, cancelled, err := adapter.SelectEnvironment("Select environment", choices)
	if err != nil || cancelled || value != ".env.production" {
		t.Fatalf("SelectEnvironment() = (%q, %t, %v)", value, cancelled, err)
	}
	if err := run.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if stdout.Len() != 0 || !strings.Contains(diagnostics.String(), "invalid selection") || terminaltest.ContainsTerminalControl(diagnostics.Bytes()) {
		t.Fatalf("Plain streams = (%q, %q)", stdout.String(), diagnostics.String())
	}

	cancelledExperience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Session:     terminalexperience.Session{Kind: terminalexperience.PlainInteractive},
		Input:       strings.NewReader("quit\n"),
		Diagnostics: &bytes.Buffer{},
	})
	cancelledRun := cancelledExperience.Open(context.Background())
	cancelledAdapter := newTerminalExportEnvAdapter(cancelledRun, cancelledExperience.Session())
	value, cancelled, err = cancelledAdapter.SelectEnvironment("Select environment", choices)
	if err != nil || !cancelled || value != "" {
		t.Fatalf("cancelled SelectEnvironment() = (%q, %t, %v)", value, cancelled, err)
	}
	if err := cancelledRun.Close(); err != nil {
		t.Fatalf("cancelled Close() error = %v", err)
	}
}

func TestTerminalExportEnvAdapterPreservesAutomationResolutionAndRejectsInteraction(t *testing.T) {
	uniqueExperience := terminaltest.NewRecordingExperience()
	uniqueAdapter := newTerminalExportEnvAdapter(uniqueExperience.Open(context.Background()), terminalexperience.Session{Kind: terminalexperience.Automation})
	value, cancelled, err := uniqueAdapter.SelectEnvironment("Select environment", []EnvironmentChoice{{Value: ".env.production", Label: "production"}})
	if err != nil || cancelled || value != ".env.production" || len(uniqueExperience.Run.Operations()) != 0 {
		t.Fatalf("unique Automation selection = (%q, %t, %v), operations=%#v", value, cancelled, err, uniqueExperience.Run.Operations())
	}

	automationExperience := terminaltest.NewRecordingExperience(terminaltest.SemanticAnswer{Err: terminalexperience.ErrAutomationInteraction})
	automationAdapter := newTerminalExportEnvAdapter(automationExperience.Open(context.Background()), terminalexperience.Session{Kind: terminalexperience.Automation})
	if _, _, err := automationAdapter.SelectEnvironment("Select environment", []EnvironmentChoice{{Value: ".env", Label: "default"}, {Value: ".env.production", Label: "production"}}); !errors.Is(err, errExportEnvRequiresInteractive) {
		t.Fatalf("ambiguous Automation selection error = %v", err)
	}
}

func TestTerminalExportEnvPresentationPreservesPlainAndAutomationResults(t *testing.T) {
	for _, session := range []terminalexperience.Session{
		{Kind: terminalexperience.PlainInteractive},
		{Kind: terminalexperience.Automation},
	} {
		var output bytes.Buffer
		experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{Session: session, Output: &output})
		run := experience.Open(context.Background())
		adapter := newTerminalExportEnvAdapter(run, session)
		adapter.Outro("Exported variables:")
		adapter.Print("{\n  \"VALUE\": \"production\"\n}")
		if err := run.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
		if got, want := output.String(), "Exported variables:\n{\n  \"VALUE\": \"production\"\n}\n"; got != want {
			t.Fatalf("%v output = %q, want %q", session.Kind, got, want)
		}
		if terminaltest.ContainsTerminalControl(output.Bytes()) {
			t.Fatalf("%v output contains terminal control: %q", session.Kind, output.String())
		}
	}

	for _, testCase := range []struct {
		role terminalexperience.VisualRole
	}{
		{role: terminalexperience.VisualRoleMuted},
		{role: terminalexperience.VisualRolePlain},
		{role: terminalexperience.VisualRoleWarning},
	} {
		document := terminalExportEnvDocument(terminalexperience.Session{Kind: terminalexperience.RichInteractive, Color: true}, "result", testCase.role)
		if got := document.Blocks[0].Role; got != testCase.role {
			t.Fatalf("Rich role = %v, want %v", got, testCase.role)
		}
	}
}
