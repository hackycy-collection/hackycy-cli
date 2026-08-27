package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hackycy/hackycy-cli/internal/cliapp"
	configcm "github.com/hackycy/hackycy-cli/internal/commands/config/cm"
	"github.com/hackycy/hackycy-cli/internal/logging"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestTerminalCMRemoveAdapterTranslatesConfirmation(t *testing.T) {
	experience := terminaltest.NewRecordingExperience(terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Confirmed: true}})
	run := experience.Open(context.Background())
	adapter := newTerminalCMRemoveAdapter(run, terminalexperience.Session{Kind: terminalexperience.RichInteractive, Color: true})

	confirmed, cancelled, err := adapter.Confirm(configcm.RemoveConfirmPrompt{Message: `Remove CM profile "work"?`})

	if err != nil || cancelled || !confirmed {
		t.Fatalf("Confirm() = (%t, %t, %v)", confirmed, cancelled, err)
	}
	if err := run.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	operations := experience.Run.Operations()
	if len(operations) != 2 || operations[0].Kind != terminaltest.AskOperation || operations[1].Kind != terminaltest.CloseOperation {
		t.Fatalf("operations = %#v", operations)
	}
	request := operations[0].Value.(terminalexperience.InteractionRequest)
	if request.Kind != terminalexperience.InteractionConfirm || request.Message != `Remove CM profile "work"?` || !request.HasDefault || request.Default.Confirmed {
		t.Fatalf("confirmation request = %#v", request)
	}
}

func TestTerminalCMRemoveAdapterMapsCancellationAndAutomation(t *testing.T) {
	cancelledExperience := terminaltest.NewRecordingExperience(terminaltest.SemanticAnswer{Err: terminalexperience.ErrInteractionCancelled})
	cancelledAdapter := newTerminalCMRemoveAdapter(cancelledExperience.Open(context.Background()), terminalexperience.Session{Kind: terminalexperience.RichInteractive})
	confirmed, cancelled, err := cancelledAdapter.Confirm(configcm.RemoveConfirmPrompt{Message: "Remove CM profile \"work\"?"})
	if err != nil || !cancelled || confirmed {
		t.Fatalf("cancelled Confirm() = (%t, %t, %v)", confirmed, cancelled, err)
	}

	automationExperience := terminaltest.NewRecordingExperience(terminaltest.SemanticAnswer{Err: terminalexperience.ErrAutomationInteraction})
	automationAdapter := newTerminalCMRemoveAdapter(automationExperience.Open(context.Background()), terminalexperience.Session{Kind: terminalexperience.Automation})
	if _, _, err := automationAdapter.Confirm(configcm.RemoveConfirmPrompt{Message: "Remove CM profile \"work\"?"}); !errors.Is(err, errConfigCMRemoveRequiresInteractive) {
		t.Fatalf("Automation Confirm() error = %v", err)
	}
}

func TestTerminalCMRemovePresentationUsesTheSharedOutputBoundary(t *testing.T) {
	var output bytes.Buffer
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Session: terminalexperience.Session{Kind: terminalexperience.PlainInteractive},
		Output:  &output,
	})
	run := experience.Open(context.Background())
	adapter := newTerminalCMRemoveAdapter(run, experience.Session())
	adapter.Cancel("Cancelled")
	adapter.Success("Profile work removed")
	if err := run.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if got, want := output.String(), "Cancelled\nProfile work removed\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
	if terminaltest.ContainsTerminalControl(output.Bytes()) {
		t.Fatalf("plain output contains terminal control: %q", output.String())
	}
	for _, testCase := range []struct {
		cancelled bool
		role      terminalexperience.VisualRole
	}{
		{cancelled: true, role: terminalexperience.VisualRoleWarning},
		{role: terminalexperience.VisualRoleSuccess},
	} {
		document := terminalCMRemoveDocument(terminalexperience.Session{Kind: terminalexperience.RichInteractive, Color: true}, "result", testCase.cancelled)
		if got := document.Blocks[0].Role; got != testCase.role {
			t.Fatalf("Rich role = %v, want %v", got, testCase.role)
		}
	}
}

func TestConfigCMRemoveAutomationValidatesBeforePromptAndMutation(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		profile   string
		wantError string
	}{
		{name: "missing profile keeps validation error", profile: "missing", wantError: "error: CM profile not found: missing\n"},
		{name: "valid profile requires interaction", profile: "work", wantError: "error: config cm remove requires an interactive terminal\n"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", "")
			configPath := writeCMRemoveConfig(t, home)
			before, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
			experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
				Session:     terminalexperience.Session{Kind: terminalexperience.Automation},
				Input:       panicCMRemoveReader{},
				Output:      stdout,
				Diagnostics: stderr,
			})
			app, err := cliapp.New(cliapp.BuildInfo{Version: "0.0.0-dev"}, cliapp.Dependencies{
				Out:            stdout,
				Err:            stderr,
				Logging:        logging.NewRuntime(logging.Options{Writer: stderr}),
				ConfigCMRemove: newConfigCMRemoveHandler(experience),
			})
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			outcome := app.Execute(context.Background(), []string{"config", "cm", "remove", testCase.profile})
			if outcome.Code != 1 || stdout.Len() != 0 || stderr.String() != testCase.wantError || terminaltest.ContainsTerminalControl(append(append([]byte{}, stdout.Bytes()...), stderr.Bytes()...)) {
				t.Fatalf("Automation outcome = %#v, stdout = %q, stderr = %q", outcome, stdout.String(), stderr.String())
			}
			if testCase.profile == "work" && !errors.Is(outcome.Err, errConfigCMRemoveRequiresInteractive) {
				t.Fatalf("valid Automation error = %v", outcome.Err)
			}
			after, err := os.ReadFile(configPath)
			if err != nil || !bytes.Equal(after, before) {
				t.Fatalf("Automation failure changed configuration = (%v, %q)", err, after)
			}
		})
	}
}

func TestConfigCMRemovePlainConfirmationAndCancellationPreserveMutationBoundary(t *testing.T) {
	for _, testCase := range []struct {
		name           string
		input          string
		wantOutput     string
		wantResult     configcm.RemoveResult
		wantDiagnostic string
		removed        bool
	}{
		{name: "confirmed", input: "y\n", wantOutput: "Profile work removed\n", removed: true},
		{name: "declined", input: "\n", wantOutput: "Cancelled\n", wantResult: configcm.RemoveResult{Declined: true}},
		{name: "declined after invalid confirmation", input: "maybe\n\n", wantOutput: "Cancelled\n", wantResult: configcm.RemoveResult{Declined: true}, wantDiagnostic: "please answer yes or no"},
		{name: "cancelled", input: "", wantOutput: "Cancelled\n", wantResult: configcm.RemoveResult{Cancelled: true}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", "")
			configPath := writeCMRemoveConfig(t, home)
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

			result, err := newConfigCMRemoveHandler(experience)(context.Background(), configcm.RemoveRequest{Profile: "work"})
			if err != nil || result != testCase.wantResult {
				t.Fatalf("Remove() = (%#v, %v), want (%#v, nil)", result, err, testCase.wantResult)
			}
			if got := stdout.String(); got != testCase.wantOutput {
				t.Fatalf("stdout = %q, want %q", got, testCase.wantOutput)
			}
			if terminaltest.ContainsTerminalControl(append(append([]byte{}, stdout.Bytes()...), diagnostics.Bytes()...)) {
				t.Fatalf("Plain output contains terminal control: (%q, %q)", stdout.String(), diagnostics.String())
			}
			if testCase.wantDiagnostic != "" && !strings.Contains(diagnostics.String(), testCase.wantDiagnostic) {
				t.Fatalf("diagnostics = %q, missing %q", diagnostics.String(), testCase.wantDiagnostic)
			}
			after, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatalf("read result configuration: %v", err)
			}
			if !testCase.removed {
				if !bytes.Equal(after, before) {
					t.Fatalf("non-mutating result changed configuration: %q", after)
				}
				return
			}
			document := standaloneCMConfig(t, after)
			if document.CM.DefaultProfile != "keep" || len(document.CM.Profiles) != 1 {
				t.Fatalf("confirmed removal = %#v", document.CM)
			}
			if _, found := document.CM.Profiles["work"]; found {
				t.Fatalf("confirmed removal retained work profile: %#v", document.CM)
			}
		})
	}
}

func TestConfigCMRemoveStandaloneBinary(t *testing.T) {
	root := repositoryRoot(t)
	binary := standaloneBinaryOutputPath(filepath.Join(t.TempDir(), "ycy"))
	build := exec.Command("go", "build", "-trimpath", "-o", binary, "./cmd/ycy")
	build.Dir = root
	build.Env = environmentWith(map[string]string{
		"CGO_ENABLED": "0",
		"GOTOOLCHAIN": "go1.26.7",
		"GOWORK":      "off",
	})
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build standalone binary: %v\n%s", err, output)
	}

	home := t.TempDir()
	configPath := writeCMRemoveConfig(t, home)
	environment := environmentWith(map[string]string{"HOME": home, "USERPROFILE": ""})
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	output, err := runStandaloneWithInput(binary, environment, "yes\n", "config", "cm", "remove", "work")
	if err == nil || string(output) != "error: config cm remove requires an interactive terminal\n" {
		t.Fatalf("valid Automation removal = (%v, %q)", err, output)
	}
	after, err := os.ReadFile(configPath)
	if err != nil || !bytes.Equal(after, before) {
		t.Fatalf("valid Automation removal changed config = (%v, %q)", err, after)
	}

	output, err = runStandaloneWithInput(binary, environment, "yes\n", "config", "cm", "remove", "missing")
	if err == nil || string(output) != "error: CM profile not found: missing\n" {
		t.Fatalf("missing Automation removal = (%v, %q)", err, output)
	}
	after, err = os.ReadFile(configPath)
	if err != nil || !bytes.Equal(after, before) {
		t.Fatalf("missing Automation removal changed config = (%v, %q)", err, after)
	}

	helpOutput, err := runStandalone(binary, environment, "config", "cm", "--help")
	if err != nil || !strings.Contains(string(helpOutput), "list") || !strings.Contains(string(helpOutput), "add") || !strings.Contains(string(helpOutput), "use") || !strings.Contains(string(helpOutput), "set") || !strings.Contains(string(helpOutput), "remove") || !strings.Contains(string(helpOutput), "test") {
		t.Fatalf("cm help = (%v, %q)", err, helpOutput)
	}
}

func writeCMRemoveConfig(t *testing.T, home string) string {
	t.Helper()
	directory := filepath.Join(home, ".ycy-cli")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatalf("create config directory: %v", err)
	}
	path := filepath.Join(directory, "config.json")
	contents := `{
  "salt": "bGVnYWN5LWNvbmZpZy1zYWx0",
  "fork": {"instances": {"github": {"host": "github.com", "type": "github", "token": "fork-token"}}},
  "cm": {"defaultProfile": "work", "profiles": {
    "work": {"baseURL": "https://work.example/v1", "model": "work-model", "apiKey": "work-api-key-that-must-not-escape"},
    "keep": {"baseURL": "https://keep.example/v1", "model": "keep-model", "apiKey": "keep-api-key-that-must-not-escape"}
  }}
}`
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}
	return path
}

type panicCMRemoveReader struct{}

func (panicCMRemoveReader) Read([]byte) (int, error) {
	panic("config cm remove attempted to read Automation input")
}
