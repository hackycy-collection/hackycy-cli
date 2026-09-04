package remove

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestConfigCMRemoveConsoleDescriptorProvidesSafeBoundedContext(t *testing.T) {
	want := terminalexperience.ConsoleDescriptor{
		Command: "YCY / config cm remove",
		Target:  "commit message profile removal",
		Status:  "READY",
		Metadata: []terminalexperience.ConsoleMetadata{
			{Label: "scope", Value: "commit message configuration"},
			{Label: "profile", Value: "work"},
		},
	}
	if got := terminalCMRemoveConsoleDescriptor("work"); !reflect.DeepEqual(got, want) {
		t.Fatalf("Console descriptor = %#v, want %#v", got, want)
	}
	unsafe := terminalCMRemoveConsoleDescriptor("bad\nprofile")
	for _, field := range []string{unsafe.Command, unsafe.Target, unsafe.Status} {
		if terminaltest.ContainsTerminalControl([]byte(field)) {
			t.Fatalf("descriptor field contains terminal control: %q", field)
		}
	}
	for _, metadata := range unsafe.Metadata {
		if terminaltest.ContainsTerminalControl([]byte(metadata.Label)) || terminaltest.ContainsTerminalControl([]byte(metadata.Value)) {
			t.Fatalf("descriptor metadata contains terminal control: %#v", metadata)
		}
	}
	if got := unsafe.Metadata[1].Value; got != "Profile configured" {
		t.Fatalf("unsafe profile projection = %q, want Profile configured", got)
	}
}

func TestTerminalCMRemoveAdapterTranslatesConfirmation(t *testing.T) {
	experience := terminaltest.NewRecordingExperience(terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Confirmed: true}})
	run := experience.Open(context.Background())
	adapter := newTerminalCMRemoveAdapter(run)

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
	cancelledAdapter := newTerminalCMRemoveAdapter(cancelledExperience.Open(context.Background()))
	confirmed, cancelled, err := cancelledAdapter.Confirm(RemoveConfirmPrompt{Message: "Remove CM profile \"work\"?"})
	if err != nil || !cancelled || confirmed {
		t.Fatalf("cancelled Confirm() = (%t, %t, %v)", confirmed, cancelled, err)
	}

	automationExperience := terminaltest.NewRecordingExperience(terminaltest.SemanticAnswer{Err: terminalexperience.ErrAutomationInteraction})
	automationAdapter := newTerminalCMRemoveAdapter(automationExperience.Open(context.Background()))
	if _, _, err := automationAdapter.Confirm(RemoveConfirmPrompt{Message: "Remove CM profile \"work\"?"}); !errors.Is(err, errConfigCMRemoveRequiresInteractive) {
		t.Fatalf("Automation Confirm() error = %v", err)
	}
	contextExperience := terminaltest.NewRecordingExperience(terminaltest.SemanticAnswer{Err: context.Canceled})
	contextAdapter := newTerminalCMRemoveAdapter(contextExperience.Open(context.Background()))
	if _, cancelled, err := contextAdapter.Confirm(RemoveConfirmPrompt{Message: "Remove CM profile \"work\"?"}); !errors.Is(err, context.Canceled) || cancelled {
		t.Fatalf("context Confirm() = (cancelled=%t, err=%v), want original context error", cancelled, err)
	}
}

func TestTerminalCMRemovePresentationUsesTheSharedOutputBoundary(t *testing.T) {
	var output bytes.Buffer
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Capabilities: terminalexperience.Capabilities{Interaction: terminalexperience.PlainInteractive},
		Output:       &output,
	})
	run := experience.Open(context.Background())
	adapter := newTerminalCMRemoveAdapter(run)
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
		document := terminalCMRemoveDocument("result", testCase.cancelled)
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
				Capabilities: terminalexperience.Capabilities{Interaction: terminalexperience.Automation},
				Input:        panicCMRemoveReader{},
				Output:       stdout,
				Diagnostics:  stderr,
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
				Capabilities: terminalexperience.Capabilities{Interaction: terminalexperience.PlainInteractive},
				Input:        strings.NewReader(testCase.input),
				Output:       stdout,
				Diagnostics:  diagnostics,
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

func TestConfigCMRemoveContextCancellationDuringValidationPreservesError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var writes int
	var output, diagnostics bytes.Buffer
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Capabilities: terminalexperience.Capabilities{Interaction: terminalexperience.PlainInteractive},
		Input:        panicCMRemoveReader{},
		Output:       &output,
		Diagnostics:  &diagnostics,
	})
	options := &Options{
		Context:  ctx,
		Profile:  "work",
		Terminal: experience,
		Store: func() (Reader, RemoveWriter, error) {
			return cmRemoveReaderFunc(func() (appconfig.CMProfileList, error) {
					cancel()
					return appconfig.CMProfileList{DefaultProfile: "work", Profiles: []appconfig.CMProfile{{Name: "work"}}}, nil
				}), cmRemoveWriterFunc(func(string) (bool, error) {
					writes++
					return true, nil
				}), nil
		},
	}
	_, err := executeRemove(options)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("executeRemove() error = %v, want context cancellation", err)
	}
	if writes != 0 || output.Len() != 0 {
		t.Fatalf("context cancellation wrote state/result: writes=%d stdout=%q", writes, output.String())
	}
}

func TestCMRemovePhaseSinkTracksValidationAndRemovalSeparately(t *testing.T) {
	experience := terminaltest.NewRecordingExperience()
	run := experience.Open(context.Background())
	sink := newCMRemovePhaseSink(run, terminalexperience.Capabilities{Interaction: terminalexperience.RichInteractive})
	sink.beginValidation()
	sink.endValidation(terminalexperience.PhaseCompleted, "safe validation")
	sink.beginRemoval()
	sink.endRemoval(terminalexperience.PhaseCompleted, "Profile removed")

	operations := experience.Run.Operations()
	if len(operations) != 2 || operations[0].Kind != terminaltest.TrackOperation || operations[1].Kind != terminaltest.TrackOperation {
		t.Fatalf("operations = %#v", operations)
	}
	first := operations[0].Value.(terminalexperience.TrackedOperation)
	second := operations[1].Value.(terminalexperience.TrackedOperation)
	wantCatalog := []terminalexperience.PhaseDefinition{{ID: cmRemoveValidationPhaseID, Name: cmRemoveValidationPhaseName}, {ID: cmRemovePhaseID, Name: cmRemovePhaseName}}
	if !reflect.DeepEqual(first.Phases, wantCatalog) || !reflect.DeepEqual(second.Phases, wantCatalog) {
		t.Fatalf("phase catalogs = %#v / %#v", first.Phases, second.Phases)
	}
	var firstUpdates, secondUpdates []terminalexperience.OperationPhase
	for update := range first.Updates {
		firstUpdates = append(firstUpdates, update)
	}
	for update := range second.Updates {
		secondUpdates = append(secondUpdates, update)
	}
	if !reflect.DeepEqual(firstUpdates, []terminalexperience.OperationPhase{{ID: cmRemoveValidationPhaseID, State: terminalexperience.PhaseActive, Detail: "Checking profile"}, {ID: cmRemoveValidationPhaseID, State: terminalexperience.PhaseCompleted, Detail: "safe validation"}}) {
		t.Fatalf("validation updates = %#v", firstUpdates)
	}
	if !reflect.DeepEqual(secondUpdates, []terminalexperience.OperationPhase{{ID: cmRemovePhaseID, State: terminalexperience.PhaseActive, Detail: "Deleting stored profile"}, {ID: cmRemovePhaseID, State: terminalexperience.PhaseCompleted, Detail: "Profile removed"}}) {
		t.Fatalf("removal updates = %#v", secondUpdates)
	}
}

type cmRemoveReaderFunc func() (appconfig.CMProfileList, error)

func (function cmRemoveReaderFunc) ListCMProfiles() (appconfig.CMProfileList, error) {
	return function()
}

type cmRemoveWriterFunc func(string) (bool, error)

func (function cmRemoveWriterFunc) RemoveCMProfile(name string) (bool, error) {
	return function(name)
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
