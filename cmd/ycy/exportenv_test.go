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

	"github.com/hackycy/hackycy-cli/internal/commands/exportenv"
	"github.com/hackycy/hackycy-cli/internal/logging"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestTerminalExportEnvAdapterTranslatesSelectionAndPresentation(t *testing.T) {
	experience := terminaltest.NewRecordingExperience(terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Value: ".env.production"}})
	run := experience.Open(context.Background())
	adapter := newTerminalExportEnvAdapter(run, terminalexperience.Session{Kind: terminalexperience.RichInteractive, Color: true})
	choices := []exportenv.EnvironmentChoice{
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
	choices := []exportenv.EnvironmentChoice{
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
	value, cancelled, err := uniqueAdapter.SelectEnvironment("Select environment", []exportenv.EnvironmentChoice{{Value: ".env.production", Label: "production"}})
	if err != nil || cancelled || value != ".env.production" || len(uniqueExperience.Run.Operations()) != 0 {
		t.Fatalf("unique Automation selection = (%q, %t, %v), operations=%#v", value, cancelled, err, uniqueExperience.Run.Operations())
	}

	automationExperience := terminaltest.NewRecordingExperience(terminaltest.SemanticAnswer{Err: terminalexperience.ErrAutomationInteraction})
	automationAdapter := newTerminalExportEnvAdapter(automationExperience.Open(context.Background()), terminalexperience.Session{Kind: terminalexperience.Automation})
	if _, _, err := automationAdapter.SelectEnvironment("Select environment", []exportenv.EnvironmentChoice{{Value: ".env", Label: "default"}, {Value: ".env.production", Label: "production"}}); !errors.Is(err, errExportEnvRequiresInteractive) {
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

func TestExportEnvAutomationPreservesResolvedPathsAndRejectsAmbiguityBeforeEffects(t *testing.T) {
	workingDirectory := t.TempDir()
	writeExportEnvFile(t, filepath.Join(workingDirectory, "named", ".env.production"), "VALUE=production\n")
	writeExportEnvFile(t, filepath.Join(workingDirectory, "unique", ".env.production"), "VALUE=unique\n")
	writeExportEnvFile(t, filepath.Join(workingDirectory, "ambiguous", ".env"), "BASE=base\n")
	writeExportEnvFile(t, filepath.Join(workingDirectory, "ambiguous", ".env.production"), "VALUE=production\n")
	protectedPath := filepath.Join(workingDirectory, "protected.json")
	if err := os.WriteFile(protectedPath, []byte("protected"), 0o600); err != nil {
		t.Fatalf("write protected output: %v", err)
	}
	withExportEnvWorkingDirectory(t, workingDirectory)

	for _, testCase := range []struct {
		name      string
		arguments []string
		wantOut   string
		wantErr   string
	}{
		{name: "named environment", arguments: []string{"export", "env", "named", "--env", "production"}, wantOut: "Exported variables:\n{\n  \"VALUE\": \"production\"\n}\n"},
		{name: "unique environment", arguments: []string{"export", "env", "unique"}, wantOut: "Exported variables:\n{\n  \"VALUE\": \"unique\"\n}\n"},
		{name: "ambiguous environment", arguments: []string{"export", "env", "ambiguous", "--out", "protected.json"}, wantErr: "error: export env requires an interactive terminal\n"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
			experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
				Session:     terminalexperience.Session{Kind: terminalexperience.Automation},
				Input:       panicExportEnvReader{},
				Output:      stdout,
				Diagnostics: stderr,
			})
			app, err := newRootCommandForTest("0.0.0-dev", rootTestDependencies{
				Out:       stdout,
				Err:       stderr,
				Logging:   logging.NewRuntime(logging.Options{Writer: stderr}),
				ExportEnv: newExportEnvHandler(experience),
			})
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			outcome := app.Execute(context.Background(), testCase.arguments)
			if testCase.wantErr == "" {
				if outcome.Code != 0 || outcome.Err != nil || stdout.String() != testCase.wantOut || stderr.Len() != 0 {
					t.Fatalf("resolved Automation outcome = %#v, stdout = %q, stderr = %q", outcome, stdout.String(), stderr.String())
				}
			} else if outcome.Code != 1 || !errors.Is(outcome.Err, errExportEnvRequiresInteractive) || stdout.Len() != 0 || stderr.String() != testCase.wantErr {
				t.Fatalf("ambiguous Automation outcome = %#v, stdout = %q, stderr = %q", outcome, stdout.String(), stderr.String())
			}
			if terminaltest.ContainsTerminalControl(append(append([]byte{}, stdout.Bytes()...), stderr.Bytes()...)) {
				t.Fatalf("Automation streams contain terminal control: (%q, %q)", stdout.String(), stderr.String())
			}
		})
	}
	contents, err := os.ReadFile(protectedPath)
	if err != nil || string(contents) != "protected" {
		t.Fatalf("ambiguous Automation changed output target = (%v, %q)", err, contents)
	}
}

func TestExportEnvPlainCancellationDoesNotWriteOutput(t *testing.T) {
	workingDirectory := t.TempDir()
	writeExportEnvFile(t, filepath.Join(workingDirectory, "project", ".env"), "BASE=base\n")
	writeExportEnvFile(t, filepath.Join(workingDirectory, "project", ".env.production"), "VALUE=production\n")
	protectedPath := filepath.Join(workingDirectory, "protected.json")
	if err := os.WriteFile(protectedPath, []byte("protected"), 0o600); err != nil {
		t.Fatalf("write protected output: %v", err)
	}
	withExportEnvWorkingDirectory(t, workingDirectory)
	stdout, diagnostics := &bytes.Buffer{}, &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Session:     terminalexperience.Session{Kind: terminalexperience.PlainInteractive},
		Input:       strings.NewReader("cancel\n"),
		Output:      stdout,
		Diagnostics: diagnostics,
	})
	app, err := newRootCommandForTest("0.0.0-dev", rootTestDependencies{
		Out:       stdout,
		Err:       diagnostics,
		Logging:   logging.NewRuntime(logging.Options{Writer: diagnostics}),
		ExportEnv: newExportEnvHandler(experience),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	outcome := app.Execute(context.Background(), []string{"export", "env", "project", "--out", "protected.json"})
	if outcome.Code != 0 || outcome.Err != nil || stdout.String() != "Cancelled\n" || terminaltest.ContainsTerminalControl(append(stdout.Bytes(), diagnostics.Bytes()...)) {
		t.Fatalf("Plain cancellation outcome = %#v, stdout = %q, diagnostics = %q", outcome, stdout.String(), diagnostics.String())
	}
	contents, err := os.ReadFile(protectedPath)
	if err != nil || string(contents) != "protected" {
		t.Fatalf("Plain cancellation changed output target = (%v, %q)", err, contents)
	}
}

func writeExportEnvFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("create %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func withExportEnvWorkingDirectory(t *testing.T, directory string) {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir(directory); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
}

type panicExportEnvReader struct{}

func (panicExportEnvReader) Read([]byte) (int, error) {
	panic("export env attempted to read Automation input")
}
