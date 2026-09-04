package add

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
	adapter := newTerminalCMAddAdapter(run)

	input, cancelled, err := PromptAdd(adapter)
	if err != nil || cancelled {
		t.Fatalf("PromptAdd() = (%#v, %t, %v)", input, cancelled, err)
	}
	if got, want := input, (AddInput{Name: "work", BaseURL: "https://provider.example/v1", Model: "gpt-4.1-mini", APIKey: "secret-api-key"}); got != want {
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
		if request.TranscriptProject == nil && request.Kind != terminalexperience.InteractionSecret {
			t.Fatalf("request %d has no safe transcript projection", index)
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
	nameRequest := operations[0].Value.(terminalexperience.InteractionRequest)
	if got := nameRequest.TranscriptProject(terminalexperience.InteractionAnswer{Value: "safe-name"}); got != "safe-name" {
		t.Fatalf("name transcript = %q", got)
	}
	urlRequest := operations[1].Value.(terminalexperience.InteractionRequest)
	if got := urlRequest.TranscriptProject(terminalexperience.InteractionAnswer{Value: "https://user:pass@example.test/v1?token=hidden#frag"}); got != "https://example.test/v1" {
		t.Fatalf("URL transcript = %q", got)
	}
	modelRequest := operations[2].Value.(terminalexperience.InteractionRequest)
	if got := modelRequest.TranscriptProject(terminalexperience.InteractionAnswer{Value: "bad\nmodel"}); got != "Model configured" {
		t.Fatalf("model transcript = %q", got)
	}
}

func TestConfigCMAddConsoleDescriptorProvidesSafeBoundedContext(t *testing.T) {
	want := terminalexperience.ConsoleDescriptor{
		Command: "YCY / config cm add",
		Target:  "commit message profile setup",
		Status:  "READY",
		Metadata: []terminalexperience.ConsoleMetadata{{
			Label: "scope",
			Value: "commit message configuration",
		}},
	}
	if got := terminalCMAddConsoleDescriptor(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Console descriptor = %#v, want %#v", got, want)
	}
}

func TestTerminalCMAddAdapterMapsTerminalCancellation(t *testing.T) {
	experience := terminaltest.NewRecordingExperience(terminaltest.SemanticAnswer{Err: terminalexperience.ErrInteractionCancelled})
	adapter := newTerminalCMAddAdapter(experience.Open(context.Background()))

	input, cancelled, err := PromptAdd(adapter)
	if err != nil || !cancelled || input != (AddInput{}) {
		t.Fatalf("PromptAdd() = (%#v, %t, %v)", input, cancelled, err)
	}
}

func TestTerminalCMAddAdapterRoutesPlainPromptAndValidationToDiagnostics(t *testing.T) {
	stdout, diagnostics := &bytes.Buffer{}, &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Capabilities: terminalexperience.Capabilities{Interaction: terminalexperience.PlainInteractive},
		Input:        strings.NewReader("\nwork\n"),
		Output:       stdout,
		Diagnostics:  diagnostics,
	})
	run := experience.Open(context.Background())
	adapter := newTerminalCMAddAdapter(run)
	value, cancelled, err := adapter.Text(AddTextPrompt{
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
		Capabilities: terminalexperience.Capabilities{Interaction: terminalexperience.PlainInteractive},
		Output:       &output,
	})
	run := experience.Open(context.Background())
	adapter := newTerminalCMAddAdapter(run)
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
		document := terminalCMAddDocument("result", testCase.cancelled)
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
		Capabilities: terminalexperience.Capabilities{Interaction: terminalexperience.Automation},
		Input:        panicCMAddReader{},
		Output:       stdout,
		Diagnostics:  stderr,
	})
	err := runAdd(&Options{
		Context:  context.Background(),
		Terminal: experience,
		Store: func() (AddWriter, error) {
			panic("config cm add attempted to construct the store")
		},
	})
	if !errors.Is(err, errConfigCMAddRequiresInteractive) {
		t.Fatalf("runAdd() error = %v", err)
	}
	if stdout.Len() != 0 || stderr.Len() != 0 || terminaltest.ContainsTerminalControl(stderr.Bytes()) {
		t.Fatalf("Automation streams = (%q, %q)", stdout.String(), stderr.String())
	}
	if _, err := os.Stat(filepath.Join(home, ".ycy-cli", "config.json")); !os.IsNotExist(err) {
		t.Fatalf("Automation failure wrote configuration: %v", err)
	}
}

func TestRunCMAddCancellationAfterFormDoesNotWrite(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reader := &cancelAfterCMAddLines{reader: strings.NewReader("work\nhttps://provider.example/v1\nmodel\n"), cancel: cancel, cancelAt: 3}
	var output, diagnostics bytes.Buffer
	var writes int
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Capabilities: terminalexperience.Capabilities{Interaction: terminalexperience.PlainInteractive},
		Input:        reader,
		Output:       &output,
		Diagnostics:  &diagnostics,
	})
	err := runAdd(&Options{
		Context:  ctx,
		Terminal: experience,
		Store: func() (AddWriter, error) {
			return cmAddWriterFunc(func(string, string, string, string) error {
				writes++
				return nil
			}), nil
		},
	})
	if err != nil {
		t.Fatalf("runAdd() error = %v, want interactive cancellation", err)
	}
	if writes != 0 {
		t.Fatalf("writes = %d, want 0", writes)
	}
	if got, want := output.String(), "Cancelled\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
	if terminaltest.ContainsTerminalControl(append(output.Bytes(), diagnostics.Bytes()...)) {
		t.Fatalf("cancellation streams contain terminal controls: stdout=%q diagnostics=%q", output.String(), diagnostics.String())
	}
}

func TestCMAddPhaseSinkReplaysCollectAndSaveStates(t *testing.T) {
	experience := terminaltest.NewRecordingExperience()
	run := experience.Open(context.Background())
	sink := newCMAddPhaseSink(run, terminalexperience.Capabilities{Interaction: terminalexperience.RichInteractive})
	sink.beginCollect()
	sink.endCollect(terminalexperience.PhaseCompleted, "safe summary")
	sink.beginSave()
	sink.endSave(terminalexperience.PhaseCompleted, "Profile saved")

	operations := experience.Run.Operations()
	if len(operations) != 1 || operations[0].Kind != terminaltest.TrackOperation {
		t.Fatalf("operations = %#v", operations)
	}
	tracked := operations[0].Value.(terminalexperience.TrackedOperation)
	if got, want := tracked.Phases, []terminalexperience.PhaseDefinition{{ID: cmAddCollectPhaseID, Name: cmAddCollectPhaseName}, {ID: cmAddSavePhaseID, Name: cmAddSavePhaseName}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("phase catalog = %#v, want %#v", got, want)
	}
	var updates []terminalexperience.OperationPhase
	for update := range tracked.Updates {
		updates = append(updates, update)
	}
	want := []terminalexperience.OperationPhase{
		{ID: cmAddCollectPhaseID, State: terminalexperience.PhaseActive, Detail: "Answer the four profile fields"},
		{ID: cmAddCollectPhaseID, State: terminalexperience.PhaseCompleted, Detail: "safe summary"},
		{ID: cmAddSavePhaseID, State: terminalexperience.PhaseActive, Detail: "Writing encrypted profile"},
		{ID: cmAddSavePhaseID, State: terminalexperience.PhaseCompleted, Detail: "Profile saved"},
	}
	if !reflect.DeepEqual(updates, want) {
		t.Fatalf("phase updates = %#v, want %#v", updates, want)
	}
}

type cmAddWriterFunc func(string, string, string, string) error

func (function cmAddWriterFunc) AddCMProfile(name, baseURL, model, apiKey string) error {
	return function(name, baseURL, model, apiKey)
}

type cancelAfterCMAddLines struct {
	reader   *strings.Reader
	cancel   context.CancelFunc
	lines    int
	cancelAt int
}

func (reader *cancelAfterCMAddLines) Read(value []byte) (int, error) {
	if len(value) == 0 {
		return 0, nil
	}
	n, err := reader.reader.Read(value[:1])
	if n == 1 && value[0] == '\n' {
		reader.lines++
		if reader.lines == reader.cancelAt {
			reader.cancel()
		}
	}
	return n, err
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
