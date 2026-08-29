package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hackycy/hackycy-cli/internal/cliapp"
	configcm "github.com/hackycy/hackycy-cli/internal/commands/config/cm"
	"github.com/hackycy/hackycy-cli/internal/logging"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestTerminalCMAddAdapterTranslatesTheOrderedForm(t *testing.T) {
	experience := terminaltest.NewRecordingExperience(
		terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Value: "work"}},
		terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Value: "https://provider.example/v1"}},
		terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Value: "gpt-4.1-mini"}},
		terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Value: "secret-api-key"}},
	)
	run := experience.Open(context.Background())
	adapter := newTerminalCMAddAdapter(run, terminalexperience.Session{Kind: terminalexperience.RichInteractive, Color: true})

	input, cancelled, err := configcm.PromptAdd(adapter)
	if err != nil || cancelled {
		t.Fatalf("PromptAdd() = (%#v, %t, %v)", input, cancelled, err)
	}
	if got, want := input, (configcm.AddInput{Name: "work", BaseURL: "https://provider.example/v1", Model: "gpt-4.1-mini", APIKey: "secret-api-key"}); got != want {
		t.Fatalf("PromptAdd() = %#v, want %#v", got, want)
	}
	if err := run.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	operations := experience.Run.Operations()
	if len(operations) != 5 {
		t.Fatalf("operations = %#v", operations)
	}
	wantKinds := []terminalexperience.InteractionKind{
		terminalexperience.InteractionText,
		terminalexperience.InteractionText,
		terminalexperience.InteractionText,
		terminalexperience.InteractionSecret,
	}
	wantMessages := []string{"Profile name", "OpenAI-compatible base URL", "Model", "API key"}
	for index := range wantKinds {
		if operations[index].Kind != terminaltest.AskOperation {
			t.Fatalf("operation %d = %#v", index, operations[index])
		}
		request := operations[index].Value.(terminalexperience.InteractionRequest)
		if request.Kind != wantKinds[index] || request.Message != wantMessages[index] {
			t.Fatalf("request %d = %#v", index, request)
		}
	}
	placeholders := []string{"e.g. openai, deepseek, work", "https://api.openai.com/v1", "gpt-4.1-mini"}
	for index, placeholder := range placeholders {
		if got := operations[index].Value.(terminalexperience.InteractionRequest).Placeholder; got != placeholder {
			t.Fatalf("placeholder %d = %q, want %q", index, got, placeholder)
		}
	}
	if err := operations[0].Value.(terminalexperience.InteractionRequest).Validate(terminalexperience.InteractionAnswer{}); err == nil || err.Error() != "Name is required" {
		t.Fatalf("name validation = %v", err)
	}
	if err := operations[3].Value.(terminalexperience.InteractionRequest).Validate(terminalexperience.InteractionAnswer{}); err == nil || err.Error() != "API key is required" {
		t.Fatalf("API key validation = %v", err)
	}
	if operations[4].Kind != terminaltest.CloseOperation {
		t.Fatalf("last operation = %#v, want close", operations[4])
	}
}

func TestTerminalCMAddAdapterMapsTerminalCancellation(t *testing.T) {
	experience := terminaltest.NewRecordingExperience(terminaltest.SemanticAnswer{Err: terminalexperience.ErrInteractionCancelled})
	adapter := newTerminalCMAddAdapter(experience.Open(context.Background()), terminalexperience.Session{Kind: terminalexperience.RichInteractive})

	input, cancelled, err := configcm.PromptAdd(adapter)
	if err != nil || !cancelled || input != (configcm.AddInput{}) {
		t.Fatalf("PromptAdd() = (%#v, %t, %v)", input, cancelled, err)
	}
}

func TestTerminalCMAddAdapterRoutesPlainPromptAndValidationToDiagnostics(t *testing.T) {
	stdout, diagnostics := &bytes.Buffer{}, &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Session:     terminalexperience.Session{Kind: terminalexperience.PlainInteractive},
		Input:       strings.NewReader("\nwork\n"),
		Output:      stdout,
		Diagnostics: diagnostics,
	})
	run := experience.Open(context.Background())
	adapter := newTerminalCMAddAdapter(run, experience.Session())
	value, cancelled, err := adapter.Text(configcm.AddTextPrompt{
		Message:     "Profile name",
		Placeholder: "e.g. openai, deepseek, work",
		Validate: func(value string) error {
			if value == "" {
				return errors.New("Name is required")
			}
			return nil
		},
	})
	if err != nil || cancelled || value != "work" {
		t.Fatalf("Text() = (%q, %t, %v)", value, cancelled, err)
	}
	if err := run.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want no prompt output", stdout.String())
	}
	for _, expected := range []string{"Profile name", "e.g. openai, deepseek, work", "Name is required"} {
		if !strings.Contains(diagnostics.String(), expected) {
			t.Fatalf("diagnostics = %q, missing %q", diagnostics.String(), expected)
		}
	}
	if terminaltest.ContainsTerminalControl(diagnostics.Bytes()) {
		t.Fatalf("Plain prompt diagnostics contain terminal control: %q", diagnostics.String())
	}
}

func TestTerminalCMAddPresentationUsesTheSharedOutputBoundary(t *testing.T) {
	var output bytes.Buffer
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Session: terminalexperience.Session{Kind: terminalexperience.PlainInteractive},
		Output:  &output,
	})
	run := experience.Open(context.Background())
	adapter := newTerminalCMAddAdapter(run, experience.Session())
	adapter.Success("Profile work added")
	adapter.Cancel("Cancelled")
	if err := run.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if got, want := output.String(), "Profile work added\nCancelled\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
	if terminaltest.ContainsTerminalControl(output.Bytes()) {
		t.Fatalf("plain output contains terminal control: %q", output.String())
	}
	for _, testCase := range []struct {
		cancelled bool
		role      terminalexperience.VisualRole
	}{
		{role: terminalexperience.VisualRoleSuccess},
		{cancelled: true, role: terminalexperience.VisualRoleWarning},
	} {
		document := terminalCMAddDocument(terminalexperience.Session{Kind: terminalexperience.RichInteractive, Color: true}, "result", testCase.cancelled)
		if got := document.Blocks[0].Role; got != testCase.role {
			t.Fatalf("Rich role = %v, want %v", got, testCase.role)
		}
	}
}

func TestConfigCMAddAutomationFailsBeforeReadOrWrite(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", "")
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Session:     terminalexperience.Session{Kind: terminalexperience.Automation},
		Input:       panicCMAddReader{},
		Output:      stdout,
		Diagnostics: stderr,
	})
	app, err := cliapp.New(cliapp.BuildInfo{Version: "0.0.0-dev"}, cliapp.Dependencies{
		Out:         stdout,
		Err:         stderr,
		Logging:     logging.NewRuntime(logging.Options{Writer: stderr}),
		ConfigCMAdd: newConfigCMAddHandler(experience),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	outcome := app.Execute(context.Background(), []string{"config", "cm", "add"})
	if outcome.Code != 1 || !errors.Is(outcome.Err, errConfigCMAddRequiresInteractive) {
		t.Fatalf("outcome = %#v", outcome)
	}
	if stdout.Len() != 0 || stderr.String() != "error: config cm add requires an interactive terminal\n" || terminaltest.ContainsTerminalControl(stderr.Bytes()) {
		t.Fatalf("Automation streams = (%q, %q)", stdout.String(), stderr.String())
	}
	if _, err := os.Stat(filepath.Join(home, ".ycy-cli", "config.json")); !os.IsNotExist(err) {
		t.Fatalf("Automation failure wrote configuration: %v", err)
	}
}

type panicCMAddReader struct{}

func (panicCMAddReader) Read([]byte) (int, error) {
	panic("config cm add attempted to read Automation input")
}

type standaloneCMConfigDocument struct {
	CM struct {
		DefaultProfile string                             `json:"defaultProfile"`
		Profiles       map[string]standaloneCMProfileData `json:"profiles"`
	} `json:"cm"`
}

type standaloneCMProfileData struct {
	BaseURL         string  `json:"baseURL"`
	Model           string  `json:"model"`
	APIKey          string  `json:"apiKey"`
	Temperature     float64 `json:"temperature"`
	TimeoutMS       int     `json:"timeoutMs"`
	MaxOutputTokens int     `json:"maxOutputTokens"`
}

func standaloneCMConfig(t *testing.T, contents []byte) standaloneCMConfigDocument {
	t.Helper()
	var document standaloneCMConfigDocument
	if err := json.Unmarshal(contents, &document); err != nil {
		t.Fatalf("decode config: %v", err)
	}
	return document
}

func standaloneCMProfile(t *testing.T, contents []byte, name string) standaloneCMProfileData {
	t.Helper()
	document := standaloneCMConfig(t, contents)
	profile, found := document.CM.Profiles[name]
	if !found {
		t.Fatalf("config omitted %q: %q", name, contents)
	}
	return profile
}
