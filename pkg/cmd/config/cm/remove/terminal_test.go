package remove

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestTerminalCMRemoveAdapterTranslatesConfirmation(t *testing.T) {
	experience := terminaltest.NewRecordingExperience(terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Confirmed: true}})
	run := experience.Open(context.Background())
	adapter := newTerminalCMRemoveAdapter(run, terminalexperience.Session{Kind: terminalexperience.RichInteractive, Color: true})

	confirmed, cancelled, err := adapter.Confirm(RemoveConfirmPrompt{Message: `Remove CM profile "work"?`})

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
	confirmed, cancelled, err := cancelledAdapter.Confirm(RemoveConfirmPrompt{Message: "Remove CM profile \"work\"?"})
	if err != nil || !cancelled || confirmed {
		t.Fatalf("cancelled Confirm() = (%t, %t, %v)", confirmed, cancelled, err)
	}

	automationExperience := terminaltest.NewRecordingExperience(terminaltest.SemanticAnswer{Err: terminalexperience.ErrAutomationInteraction})
	automationAdapter := newTerminalCMRemoveAdapter(automationExperience.Open(context.Background()), terminalexperience.Session{Kind: terminalexperience.Automation})
	if _, _, err := automationAdapter.Confirm(RemoveConfirmPrompt{Message: "Remove CM profile \"work\"?"}); !errors.Is(err, errConfigCMRemoveRequiresInteractive) {
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
			options := newCMRemoveOptions(experience)
			options.Profile = testCase.profile
			_, outcomeErr := executeRemove(options)
			wantError := strings.TrimPrefix(strings.TrimSuffix(testCase.wantError, "\n"), "error: ")
			if outcomeErr == nil || outcomeErr.Error() != wantError {
				t.Fatalf("Automation error = %v, want %q", outcomeErr, wantError)
			}
			if stdout.Len() != 0 || stderr.Len() != 0 || terminaltest.ContainsTerminalControl(append(append([]byte{}, stdout.Bytes()...), stderr.Bytes()...)) {
				t.Fatalf("Automation streams = (%q, %q)", stdout.String(), stderr.String())
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
		wantResult     RemoveResult
		wantDiagnostic string
		removed        bool
	}{
		{name: "confirmed", input: "y\n", wantOutput: "Profile work removed\n", removed: true},
		{name: "declined", input: "\n", wantOutput: "Cancelled\n", wantResult: RemoveResult{Declined: true}},
		{name: "declined after invalid confirmation", input: "maybe\n\n", wantOutput: "Cancelled\n", wantResult: RemoveResult{Declined: true}, wantDiagnostic: "please answer yes or no"},
		{name: "cancelled", input: "", wantOutput: "Cancelled\n", wantResult: RemoveResult{Cancelled: true}},
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

			options := newCMRemoveOptions(experience)
			options.Profile = "work"
			result, err := executeRemove(options)
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

func newCMRemoveOptions(experience *terminalexperience.Runtime) *Options {
	return &Options{
		Context:  context.Background(),
		Terminal: experience,
		Store: func() (Reader, RemoveWriter, error) {
			store, err := appconfig.New(appconfig.Dependencies{})
			if err != nil {
				return nil, nil, err
			}
			return store, store, nil
		},
	}
}
