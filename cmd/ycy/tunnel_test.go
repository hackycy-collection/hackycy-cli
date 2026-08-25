package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
)

func TestTerminalTunnelConnectionSelectorSelectsNumberedConnectionAndMasksTokens(t *testing.T) {
	output := &bytes.Buffer{}
	selector := newTerminalTunnelConnectionSelector(strings.NewReader("wrong\n2\n"), output)
	connections := []appconfig.TunnelConnection{
		{ID: "first", Server: "https://first.example.test", Token: "abcdefgh01234567wxyz"},
		{ID: "second", Server: "https://second.example.test", Token: "short-token"},
	}

	selected, cancelled, err := selector.Select(context.Background(), connections)
	if err != nil || cancelled || selected != "second" {
		t.Fatalf("Select() = (%q, %t, %v)", selected, cancelled, err)
	}
	text := output.String()
	for _, expected := range []string{"Select a tunnel connection", "https://first.example.test", "abcdefgh********wxyz", "***********", "Invalid selection"} {
		if !strings.Contains(text, expected) {
			t.Errorf("selector output omitted %q: %q", expected, text)
		}
	}
	if strings.Contains(text, "short-token") || strings.Contains(text, "abcdefgh01234567wxyz") {
		t.Fatalf("selector output exposed a token: %q", text)
	}
}

func TestTerminalTunnelConnectionSelectorTreatsCancellationAndEOFAsNormalCancellation(t *testing.T) {
	for _, input := range []string{"cancel\n", "unknown"} {
		t.Run(input, func(t *testing.T) {
			output := &bytes.Buffer{}
			selector := newTerminalTunnelConnectionSelector(strings.NewReader(input), output)
			selected, cancelled, err := selector.Select(context.Background(), []appconfig.TunnelConnection{{ID: "only", Server: "https://example.test", Token: "token"}})
			if err != nil || !cancelled || selected != "" || !strings.Contains(output.String(), "Cancelled") {
				t.Fatalf("Select() = (%q, %t, %v), output = %q", selected, cancelled, err, output.String())
			}
		})
	}
}

func TestTerminalTunnelConnectionSelectorForLeavesNonTTYResolutionToTheClient(t *testing.T) {
	if selector := terminalTunnelConnectionSelectorFor(strings.NewReader("1\n"), &bytes.Buffer{}); selector != nil {
		t.Fatal("non-TTY selector must remain nil so client resolution reports ambiguity")
	}
}
