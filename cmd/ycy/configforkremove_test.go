package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	configfork "github.com/hackycy/hackycy-cli/internal/commands/config/fork"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestTerminalForkRemoveAdapterTranslatesSelectionAndConfirmation(t *testing.T) {
	experience := terminaltest.NewRecordingExperience(
		terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Value: "work"}},
		terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Confirmed: true}},
	)
	run := experience.Open(context.Background())
	adapter := newTerminalForkRemoveAdapter(run, terminalexperience.Session{Kind: terminalexperience.RichInteractive, Color: true})

	selected, cancelled, err := adapter.Select(configfork.SelectPrompt{
		Message: "Select instance to remove",
		Choices: []configfork.Choice{
			{Value: "work", Label: "work (gitlab.example)"},
			{Value: "personal", Label: "personal (github.example)"},
		},
	})
	if err != nil || cancelled || selected != "work" {
		t.Fatalf("Select() = (%q, %t, %v)", selected, cancelled, err)
	}
	confirmed, cancelled, err := adapter.Confirm(configfork.ConfirmPrompt{Message: `Remove instance "work"?`})
	if err != nil || cancelled || !confirmed {
		t.Fatalf("Confirm() = (%t, %t, %v)", confirmed, cancelled, err)
	}
	if err := run.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	operations := experience.Run.Operations()
	if len(operations) != 3 || operations[0].Kind != terminaltest.AskOperation || operations[1].Kind != terminaltest.AskOperation || operations[2].Kind != terminaltest.CloseOperation {
		t.Fatalf("operations = %#v", operations)
	}
	selection := operations[0].Value.(terminalexperience.InteractionRequest)
	if selection.Kind != terminalexperience.InteractionSelect || selection.Message != "Select instance to remove" || !selection.HasDefault || selection.Default.Value != "work" || !reflect.DeepEqual(selection.Options, []terminalexperience.InteractionOption{{Label: "work (gitlab.example)", Value: "work"}, {Label: "personal (github.example)", Value: "personal"}}) {
		t.Fatalf("selection request = %#v", selection)
	}
	confirmation := operations[1].Value.(terminalexperience.InteractionRequest)
	if confirmation.Kind != terminalexperience.InteractionConfirm || confirmation.Message != `Remove instance "work"?` || !confirmation.HasDefault || confirmation.Default.Confirmed {
		t.Fatalf("confirmation request = %#v", confirmation)
	}
}

func TestTerminalForkRemoveAdapterMapsTerminalCancellation(t *testing.T) {
	experience := terminaltest.NewRecordingExperience(terminaltest.SemanticAnswer{Err: terminalexperience.ErrInteractionCancelled})
	adapter := newTerminalForkRemoveAdapter(experience.Open(context.Background()), terminalexperience.Session{Kind: terminalexperience.RichInteractive})

	value, cancelled, err := adapter.Select(configfork.SelectPrompt{Message: "Select instance to remove"})
	if err != nil || !cancelled || value != "" {
		t.Fatalf("Select() = (%q, %t, %v)", value, cancelled, err)
	}
}

func TestTerminalForkRemovePresentationUsesTheSharedOutputBoundary(t *testing.T) {
	var output bytes.Buffer
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Session: terminalexperience.Session{Kind: terminalexperience.PlainInteractive},
		Output:  &output,
	})
	run := experience.Open(context.Background())
	adapter := newTerminalForkRemoveAdapter(run, experience.Session())
	adapter.Info("No instances configured")
	adapter.Outcome("Nothing to remove")
	adapter.Outcome("Cancelled")
	adapter.Outcome("Instance work removed")
	if err := run.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if got, want := output.String(), "No instances configured\nNothing to remove\nCancelled\nInstance work removed\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
	if terminaltest.ContainsTerminalControl(output.Bytes()) {
		t.Fatalf("plain output contains terminal control: %q", output.String())
	}
	for _, testCase := range []struct {
		message string
		role    terminalexperience.VisualRole
	}{
		{message: "No instances configured", role: terminalexperience.VisualRoleMuted},
		{message: "Cancelled", role: terminalexperience.VisualRoleWarning},
		{message: "Instance work removed", role: terminalexperience.VisualRoleSuccess},
	} {
		document := terminalForkRemoveDocument(terminalexperience.Session{Kind: terminalexperience.RichInteractive}, testCase.message, testCase.role)
		if got := document.Blocks[0].Role; got != testCase.role {
			t.Fatalf("Rich role = %v, want %v", got, testCase.role)
		}
	}
}

func TestTerminalForkRemoveAdapterReportsAutomationInteraction(t *testing.T) {
	experience := terminaltest.NewRecordingExperience(terminaltest.SemanticAnswer{Err: terminalexperience.ErrAutomationInteraction})
	adapter := newTerminalForkRemoveAdapter(experience.Open(context.Background()), terminalexperience.Session{Kind: terminalexperience.Automation})
	if _, _, err := adapter.Select(configfork.SelectPrompt{Message: "Select instance to remove"}); !errors.Is(err, errConfigForkRemoveRequiresInteractive) {
		t.Fatalf("Select() error = %v", err)
	}
}

func TestConfigForkRemoveAutomationPreservesEmptyAndRejectsNonemptyBeforeMutation(t *testing.T) {
	t.Run("empty configuration", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("USERPROFILE", "")
		stdout, diagnostics := &bytes.Buffer{}, &bytes.Buffer{}
		experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
			Session:     terminalexperience.Session{Kind: terminalexperience.Automation},
			Input:       panicForkRemoveReader{},
			Output:      stdout,
			Diagnostics: diagnostics,
		})

		result, err := newConfigForkRemoveHandler(experience)(context.Background(), configfork.RemoveRequest{})
		if err != nil || result != (configfork.RemoveResult{Empty: true}) {
			t.Fatalf("Remove() = (%#v, %v)", result, err)
		}
		if got, want := stdout.String(), "No instances configured\nNothing to remove\n"; got != want {
			t.Fatalf("stdout = %q, want %q", got, want)
		}
		if diagnostics.Len() != 0 || terminaltest.ContainsTerminalControl(stdout.Bytes()) {
			t.Fatalf("Automation streams = (%q, %q)", stdout.String(), diagnostics.String())
		}
		if _, err := os.Stat(filepath.Join(home, ".ycy-cli", "config.json")); !os.IsNotExist(err) {
			t.Fatalf("empty removal created configuration: %v", err)
		}
	})

	t.Run("nonempty configuration", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("USERPROFILE", "")
		configPath := writeForkRemoveConfig(t, home)
		before, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("read fixture: %v", err)
		}
		stdout, diagnostics := &bytes.Buffer{}, &bytes.Buffer{}
		experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
			Session:     terminalexperience.Session{Kind: terminalexperience.Automation},
			Input:       panicForkRemoveReader{},
			Output:      stdout,
			Diagnostics: diagnostics,
		})

		result, err := newConfigForkRemoveHandler(experience)(context.Background(), configfork.RemoveRequest{})
		if !errors.Is(err, errConfigForkRemoveRequiresInteractive) || result != (configfork.RemoveResult{}) {
			t.Fatalf("Remove() = (%#v, %v)", result, err)
		}
		if stdout.Len() != 0 || diagnostics.Len() != 0 {
			t.Fatalf("Automation streams = (%q, %q)", stdout.String(), diagnostics.String())
		}
		after, err := os.ReadFile(configPath)
		if err != nil || !bytes.Equal(after, before) {
			t.Fatalf("Automation failure changed configuration = (%v, %q)", err, after)
		}
	})
}

func TestConfigForkRemovePlainConfirmationAndCancellationPreserveMutationBoundary(t *testing.T) {
	for _, testCase := range []struct {
		name           string
		input          string
		wantOutput     string
		wantResult     configfork.RemoveResult
		wantDiagnostic string
		removed        bool
	}{
		{name: "confirmed", input: "1\ny\n", wantOutput: "Instance work removed\n", removed: true},
		{name: "declined after invalid selection", input: "invalid\n1\n\n", wantOutput: "Cancelled\n", wantResult: configfork.RemoveResult{Declined: true}, wantDiagnostic: "invalid selection"},
		{name: "declined after invalid confirmation", input: "1\nmaybe\n\n", wantOutput: "Cancelled\n", wantResult: configfork.RemoveResult{Declined: true}, wantDiagnostic: "please answer yes or no"},
		{name: "cancelled selection", input: "", wantOutput: "Cancelled\n", wantResult: configfork.RemoveResult{Cancelled: true}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", "")
			configPath := writeForkRemoveConfig(t, home)
			before, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			stdout, diagnostics := &bytes.Buffer{}, &bytes.Buffer{}
			experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
				Session:     terminalexperience.Session{Kind: terminalexperience.PlainInteractive},
				Input:       strings.NewReader(testCase.input),
				Output:      stdout,
				Diagnostics: diagnostics,
			})

			result, err := newConfigForkRemoveHandler(experience)(context.Background(), configfork.RemoveRequest{})
			if err != nil || result != testCase.wantResult {
				t.Fatalf("Remove() = (%#v, %v), want (%#v, nil)", result, err, testCase.wantResult)
			}
			if got := stdout.String(); got != testCase.wantOutput {
				t.Fatalf("stdout = %q, want %q", got, testCase.wantOutput)
			}
			if terminaltest.ContainsTerminalControl(append(stdout.Bytes(), diagnostics.Bytes()...)) {
				t.Fatalf("Plain output contains terminal control: (%q, %q)", stdout.String(), diagnostics.String())
			}
			if testCase.wantDiagnostic != "" && !strings.Contains(diagnostics.String(), testCase.wantDiagnostic) {
				t.Fatalf("diagnostics = %q, missing %q", diagnostics.String(), testCase.wantDiagnostic)
			}
			after, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatalf("read result configuration: %v", err)
			}
			if testCase.removed {
				if bytes.Equal(after, before) || bytes.Contains(after, []byte(`"work"`)) {
					t.Fatalf("confirmed result = (%#v, %q)", result, after)
				}
				return
			}
			if !bytes.Equal(after, before) {
				t.Fatalf("non-mutating result changed configuration: %q", after)
			}
		})
	}
}

func writeForkRemoveConfig(t *testing.T, home string) string {
	t.Helper()
	directory := filepath.Join(home, ".ycy-cli")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatalf("create config directory: %v", err)
	}
	path := filepath.Join(directory, "config.json")
	contents := `{
  "salt": "bGVnYWN5LWNvbmZpZy1zYWx0",
  "fork": {"instances": {
    "work": {"host": "gitlab.example", "type": "gitlab", "token": "MDEyMzQ1Njc4OWFiY2RlZg==:tag:ciphertext"},
    "keep": {"host": "github.example", "type": "github", "token": "QWVy:second:ciphertext"}
  }}
}`
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}
	return path
}

type panicForkRemoveReader struct{}

func (panicForkRemoveReader) Read([]byte) (int, error) {
	panic("config fork remove attempted to read Automation input")
}
