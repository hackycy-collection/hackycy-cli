package add

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestTerminalForkAddAdapterTranslatesTheOrderedForm(t *testing.T) {
	experience := terminaltest.NewRecordingExperience(
		terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Value: "work"}},
		terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Value: "gitlab.example"}},
		terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Value: "gitlab"}},
		terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Value: "https"}},
		terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Value: "secret-token"}},
	)
	run := experience.Open(context.Background())
	adapter := newTerminalForkAddAdapter(run)

	input, cancelled, err := PromptAdd(adapter)
	if err != nil || cancelled {
		t.Fatalf("PromptAdd() = (%#v, %t, %v)", input, cancelled, err)
	}
	if got, want := input, (AddInput{Alias: "work", Host: "gitlab.example", Type: "gitlab", Scheme: "https", Token: "secret-token"}); got != want {
		t.Fatalf("PromptAdd() = %#v, want %#v", got, want)
	}
	if err := run.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	operations := experience.Run.Operations()
	if len(operations) != 6 {
		t.Fatalf("operations = %#v", operations)
	}
	wantKinds := []terminalexperience.InteractionKind{
		terminalexperience.InteractionText,
		terminalexperience.InteractionText,
		terminalexperience.InteractionSelect,
		terminalexperience.InteractionSelect,
		terminalexperience.InteractionSecret,
	}
	wantMessages := []string{"Instance name (alias)", "Host", "Provider type", "Protocol", "Access token"}
	for index := range wantKinds {
		if operations[index].Kind != terminaltest.AskOperation {
			t.Fatalf("operation %d kind = %q, want ask", index, operations[index].Kind)
		}
		request, ok := operations[index].Value.(terminalexperience.InteractionRequest)
		if !ok {
			t.Fatalf("operation %d request = %#v", index, operations[index].Value)
		}
		if request.Kind != wantKinds[index] || request.Message != wantMessages[index] {
			t.Fatalf("request %d = %#v", index, request)
		}
	}
	first := operations[0].Value.(terminalexperience.InteractionRequest)
	second := operations[1].Value.(terminalexperience.InteractionRequest)
	if first.Placeholder != "e.g. work, github, company-gl" || second.Placeholder != "e.g. gitlab.company.com, github.com" {
		t.Fatalf("text placeholders = (%q, %q)", first.Placeholder, second.Placeholder)
	}
	provider := operations[2].Value.(terminalexperience.InteractionRequest)
	if !provider.HasDefault || provider.Default.Value != "gitlab" || !reflect.DeepEqual(provider.Options, []terminalexperience.InteractionOption{{Label: "GitLab", Value: "gitlab"}, {Label: "GitHub", Value: "github"}}) {
		t.Fatalf("provider request = %#v", provider)
	}
	protocol := operations[3].Value.(terminalexperience.InteractionRequest)
	if !protocol.HasDefault || protocol.Default.Value != "https" || !reflect.DeepEqual(protocol.Options, []terminalexperience.InteractionOption{{Label: "HTTPS", Value: "https"}, {Label: "HTTP (self-hosted / no TLS)", Value: "http"}}) {
		t.Fatalf("protocol request = %#v", protocol)
	}
	if err := first.Validate(terminalexperience.InteractionAnswer{}); err == nil || err.Error() != "Name is required" {
		t.Fatalf("alias validation = %v", err)
	}
	if err := operations[4].Value.(terminalexperience.InteractionRequest).Validate(terminalexperience.InteractionAnswer{}); err == nil || err.Error() != "Token is required" {
		t.Fatalf("token validation = %v", err)
	}
	if operations[5].Kind != terminaltest.CloseOperation {
		t.Fatalf("last operation = %#v, want close", operations[5])
	}
}

func TestTerminalForkAddAdapterMapsTerminalCancellation(t *testing.T) {
	experience := terminaltest.NewRecordingExperience(terminaltest.SemanticAnswer{Err: terminalexperience.ErrInteractionCancelled})
	adapter := newTerminalForkAddAdapter(experience.Open(context.Background()))

	input, cancelled, err := PromptAdd(adapter)
	if err != nil || !cancelled || input != (AddInput{}) {
		t.Fatalf("PromptAdd() = (%#v, %t, %v)", input, cancelled, err)
	}
}

func TestTerminalForkAddAdapterRoutesPlainPromptAndValidationToDiagnostics(t *testing.T) {
	stdout, diagnostics := &bytes.Buffer{}, &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Capabilities: terminalexperience.Capabilities{Interaction: terminalexperience.PlainInteractive},
		Input:        strings.NewReader("\nwork\n"),
		Output:       stdout,
		Diagnostics:  diagnostics,
	})
	run := experience.Open(context.Background())
	adapter := newTerminalForkAddAdapter(run)

	value, cancelled, err := adapter.Text(TextPrompt{
		Message:     "Instance name (alias)",
		Placeholder: "e.g. work, github, company-gl",
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
	for _, expected := range []string{"Instance name (alias)", "e.g. work, github, company-gl", "Name is required"} {
		if !strings.Contains(diagnostics.String(), expected) {
			t.Fatalf("diagnostics = %q, missing %q", diagnostics.String(), expected)
		}
	}
	if terminaltest.ContainsTerminalControl(diagnostics.Bytes()) {
		t.Fatalf("Plain prompt diagnostics contain terminal control: %q", diagnostics.String())
	}
}

func TestTerminalForkAddPresentationUsesTheSharedOutputBoundary(t *testing.T) {
	var output bytes.Buffer
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Capabilities: terminalexperience.Capabilities{Interaction: terminalexperience.PlainInteractive},
		Output:       &output,
	})
	run := experience.Open(context.Background())
	adapter := newTerminalForkAddAdapter(run)
	adapter.Success("Instance work (gitlab.example) added successfully")
	adapter.Cancel("Cancelled")
	if err := run.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if got, want := output.String(), "Instance work (gitlab.example) added successfully\nCancelled\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
	if terminaltest.ContainsTerminalControl(output.Bytes()) {
		t.Fatalf("plain result contains terminal control: %q", output.String())
	}

	for _, testCase := range []struct {
		cancelled bool
		role      terminalexperience.VisualRole
	}{
		{role: terminalexperience.VisualRoleSuccess},
		{cancelled: true, role: terminalexperience.VisualRoleWarning},
	} {
		document := terminalForkAddDocument("result", testCase.cancelled)
		if got := document.Blocks[0].Role; got != testCase.role {
			t.Fatalf("Rich role = %v, want %v", got, testCase.role)
		}
	}
}

func TestConfigForkAddAutomationFailsBeforeReadOrWrite(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", "")
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Capabilities: terminalexperience.Capabilities{Interaction: terminalexperience.Automation},
		Input:        panicForkAddReader{},
		Output:       stdout,
		Diagnostics:  stderr,
	})
	err := runAdd(&Options{
		Context:  context.Background(),
		Terminal: experience,
		Store: func() (AddWriter, error) {
			panic("config fork add attempted to construct the store")
		},
	})
	if !errors.Is(err, errConfigForkAddRequiresInteractive) {
		t.Fatalf("runAdd() error = %v", err)
	}
	if got, want := stdout.String(), ""; got != want {
		t.Fatalf("stdout = %q, want empty", got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("direct leaf execution wrote diagnostics: %q", stderr.String())
	}
	if terminaltest.ContainsTerminalControl(stderr.Bytes()) {
		t.Fatalf("automation stderr contains terminal control: %q", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(home, ".ycy-cli", "config.json")); !os.IsNotExist(err) {
		t.Fatalf("Automation failure wrote configuration: %v", err)
	}
}

type panicForkAddReader struct{}

func (panicForkAddReader) Read([]byte) (int, error) {
	panic("config fork add attempted to read Automation input")
}
