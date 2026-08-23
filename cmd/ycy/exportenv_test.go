package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hackycy/hackycy-cli/internal/commands/exportenv"
)

func TestTerminalExportEnvSelectorSelectsNumberedChoice(t *testing.T) {
	output := &bytes.Buffer{}
	selector := newTerminalExportEnvSelector(strings.NewReader("2\n"), output)

	value, cancelled := selector.SelectEnvironment("Select environment", []exportenv.EnvironmentChoice{
		{Value: ".env.local", Label: "local"},
		{Value: ".env.production", Label: "production"},
	})

	if cancelled || value != ".env.production" {
		t.Fatalf("SelectEnvironment() = (%q, %t)", value, cancelled)
	}
	if !strings.Contains(output.String(), "Select environment") || !strings.Contains(output.String(), "production") {
		t.Fatalf("output = %q", output.String())
	}
}

func TestTerminalExportEnvSelectorTreatsEOFAsCancellation(t *testing.T) {
	selector := newTerminalExportEnvSelector(strings.NewReader(""), &bytes.Buffer{})

	_, cancelled := selector.SelectEnvironment("Select environment", []exportenv.EnvironmentChoice{{Value: ".env.production", Label: "production"}})

	if !cancelled {
		t.Fatal("SelectEnvironment did not report cancellation")
	}
}
