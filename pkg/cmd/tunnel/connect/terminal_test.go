package connect

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestTerminalTunnelConnectionAdapterTranslatesSelectionAndCancellation(t *testing.T) {
	connections := []appconfig.TunnelConnection{
		{ID: "first", Server: "https://first.example.test", Token: "abcdefgh01234567wxyz", LastAuthenticatedAt: "2026-01-02T03:04:05.000Z"},
		{ID: "second", Server: "https://second.example.test", Token: "short-token", LastAuthenticatedAt: "2026-01-01T03:04:05.000Z"},
	}
	experience := terminaltest.NewRecordingExperience(terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Value: "second"}})
	run := experience.Open(context.Background())
	selected, cancelled, err := newTerminalTunnelConnectionAdapter(run).Select(context.Background(), connections)
	if err != nil || cancelled || selected != "second" {
		t.Fatalf("Select() = (%q, %t, %v)", selected, cancelled, err)
	}
	if err := run.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	operations := experience.Run.Operations()
	if len(operations) != 2 || operations[0].Kind != terminaltest.AskOperation || operations[1].Kind != terminaltest.CloseOperation {
		t.Fatalf("operations = %#v", operations)
	}
	request := operations[0].Value.(terminalexperience.InteractionRequest)
	wantOptions := []terminalexperience.InteractionOption{
		{Label: "https://first.example.test  last authenticated 2026-01-02 03:04 UTC", Value: "first"},
		{Label: "https://second.example.test  last authenticated 2026-01-01 03:04 UTC", Value: "second"},
	}
	if request.Kind != terminalexperience.InteractionSelect || request.Message != "Select a tunnel connection" || request.PlainLead != "Select a tunnel connection" || request.PlainPrompt != "> " || !reflect.DeepEqual(request.Options, wantOptions) || !reflect.DeepEqual(request.CancelValues, []string{"", "q", "quit", "cancel"}) || request.ParsePlain == nil {
		t.Fatalf("selection request = %#v", request)
	}
	if strings.Contains(request.Options[0].Label, "abcdefgh01234567wxyz") || strings.Contains(request.Options[1].Label, "short-token") {
		t.Fatalf("selection options expose a token: %#v", request.Options)
	}
	if request.TranscriptLabel != "Selected tunnel connection" || request.TranscriptProject == nil || request.TranscriptProject(terminalexperience.InteractionAnswer{Value: "first"}) != wantOptions[0].Label {
		t.Fatalf("selection transcript projection = %#v", request)
	}

	cancelledExperience := terminaltest.NewRecordingExperience(terminaltest.SemanticAnswer{Err: terminalexperience.ErrInteractionCancelled})
	cancelledRun := cancelledExperience.Open(context.Background())
	selected, cancelled, err = newTerminalTunnelConnectionAdapter(cancelledRun).Select(context.Background(), connections)
	if err != nil || !cancelled || selected != "" {
		t.Fatalf("cancelled Select() = (%q, %t, %v)", selected, cancelled, err)
	}
	if err := cancelledRun.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	cancelledOperations := cancelledExperience.Run.Operations()
	if len(cancelledOperations) != 3 || cancelledOperations[1].Kind != terminaltest.MilestoneOperation || !reflect.DeepEqual(cancelledOperations[1].Value.(terminalexperience.PresentationDocument).Blocks, []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleWarning, Text: "Tunnel connection selection cancelled"}}) {
		t.Fatalf("cancelled operations = %#v", cancelledOperations)
	}
}

func TestTerminalTunnelConnectionAdapterPlainPreservesNumberedInputCancellationAndSafeProjection(t *testing.T) {
	connections := []appconfig.TunnelConnection{
		{ID: "first", Server: "https://first.example.test", Token: "abcdefgh01234567wxyz", LastAuthenticatedAt: "2026-01-02T03:04:05.000Z"},
		{ID: "second", Server: "https://second.example.test", Token: "short-token", LastAuthenticatedAt: "2026-01-01T03:04:05.000Z"},
	}
	stdout, diagnostics := &bytes.Buffer{}, &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Capabilities: terminalexperience.Capabilities{Interaction: terminalexperience.PlainInteractive},
		Input:        strings.NewReader("wrong\n2\n"),
		Output:       stdout,
		Diagnostics:  diagnostics,
	})
	run := experience.Open(context.Background())
	selected, cancelled, err := newTerminalTunnelConnectionAdapter(run).Select(context.Background(), connections)
	if err != nil || cancelled || selected != "second" || stdout.Len() != 0 || !strings.Contains(diagnostics.String(), "Invalid selection") || !strings.Contains(diagnostics.String(), "last authenticated 2026-01-02 03:04 UTC") || !strings.Contains(diagnostics.String(), "last authenticated 2026-01-01 03:04 UTC") || strings.Contains(diagnostics.String(), "abcdefgh01234567wxyz") || strings.Contains(diagnostics.String(), "short-token") || terminaltest.ContainsTerminalControl(append(stdout.Bytes(), diagnostics.Bytes()...)) {
		t.Fatalf("Plain Select() = (%q, %t, %v), streams = (%q, %q)", selected, cancelled, err, stdout.String(), diagnostics.String())
	}
	if err := run.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	for _, input := range []string{"\n", "q\n", "quit\n", "cancel\n", "unknown"} {
		t.Run(input, func(t *testing.T) {
			stdout, diagnostics := &bytes.Buffer{}, &bytes.Buffer{}
			experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
				Capabilities: terminalexperience.Capabilities{Interaction: terminalexperience.PlainInteractive},
				Input:        strings.NewReader(input),
				Output:       stdout,
				Diagnostics:  diagnostics,
			})
			run := experience.Open(context.Background())
			selected, cancelled, err := newTerminalTunnelConnectionAdapter(run).Select(context.Background(), connections)
			if err != nil || !cancelled || selected != "" || stdout.Len() != 0 || !strings.Contains(diagnostics.String(), "Tunnel connection selection cancelled") || terminaltest.ContainsTerminalControl(append(stdout.Bytes(), diagnostics.Bytes()...)) {
				t.Fatalf("Select() = (%q, %t, %v), streams = (%q, %q)", selected, cancelled, err, stdout.String(), diagnostics.String())
			}
			if err := run.Close(); err != nil {
				t.Fatalf("Close() error = %v", err)
			}
		})
	}
}

func TestTunnelConnectionInteractionOptionsUseStableOrdinalsOnlyForEqualOrigins(t *testing.T) {
	connections := []appconfig.TunnelConnection{
		{ID: "one", Server: "HTTPS://Example.test:443", Token: "token-one", LastAuthenticatedAt: "2026-02-03T00:00:00Z"},
		{ID: "two", Server: "https://example.test:443", Token: "token-two", LastAuthenticatedAt: "2026-02-02T00:00:00Z"},
		{ID: "three", Server: "https://other.test", Token: "token-three", LastAuthenticatedAt: "2026-02-01T00:00:00Z"},
	}
	options := tunnelConnectionInteractionOptions(connections)
	if len(options) != 3 || !strings.HasSuffix(options[0].Label, "(1)") || !strings.HasSuffix(options[1].Label, "(2)") || strings.Contains(options[2].Label, "(") {
		t.Fatalf("options = %#v", options)
	}
	for _, option := range options {
		if strings.Contains(option.Label, "token-") || strings.Contains(option.Label, "v1_") || strings.Contains(option.Label, "state") {
			t.Fatalf("unsafe option label = %q", option.Label)
		}
	}
}

func TestTunnelConnectionSelectorLeavesAutomationResolutionAmbiguousWithoutReadingInput(t *testing.T) {
	stdout, diagnostics := &bytes.Buffer{}, &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Capabilities: terminalexperience.Capabilities{Interaction: terminalexperience.Automation},
		Input:        panicTunnelReader{},
		Output:       stdout,
		Diagnostics:  diagnostics,
	})
	selector := terminalTunnelConnectionSelectorFor(experience)
	if selector != nil {
		t.Fatal("Automation selector must remain nil so client resolution reports ambiguity")
	}
	connections := []appconfig.TunnelConnection{
		{ID: "one", Server: "https://one.example.test", Token: "one-token"},
		{ID: "two", Server: "https://two.example.test", Token: "two-token"},
	}
	_, err := ResolveClientConfig(context.Background(), ClientOptionInput{}, ClientResolutionOptions{
		Reader:           tunnelConnectionReader{connections: connections},
		Environment:      func(string) (string, bool) { return "", false },
		SelectConnection: selector,
	})
	if err == nil || !strings.Contains(err.Error(), "Multiple remembered tunnel connections match") || stdout.Len() != 0 || diagnostics.Len() != 0 || terminaltest.ContainsTerminalControl(append(stdout.Bytes(), diagnostics.Bytes()...)) {
		t.Fatalf("Automation ambiguity = (%v), streams = (%q, %q)", err, stdout.String(), diagnostics.String())
	}
	resolved, err := ResolveClientConfig(context.Background(), ClientOptionInput{}, ClientResolutionOptions{
		Reader:      tunnelConnectionReader{connections: connections[:1]},
		Environment: func(string) (string, bool) { return "", false },
	})
	if err != nil || resolved == nil || resolved.Config.Server.String() != "https://one.example.test" || resolved.Config.Token != "one-token" {
		t.Fatalf("unique Automation resolution = (%#v, %v)", resolved, err)
	}
}

type tunnelConnectionReader struct {
	connections []appconfig.TunnelConnection
	err         error
}

func (reader tunnelConnectionReader) ReadTunnelConnections() ([]appconfig.TunnelConnection, error) {
	return append([]appconfig.TunnelConnection(nil), reader.connections...), reader.err
}

type panicTunnelReader struct{}

func (panicTunnelReader) Read([]byte) (int, error) {
	panic("Tunnel Automation must not read stdin")
}
