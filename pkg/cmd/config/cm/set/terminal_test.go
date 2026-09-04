package set

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestRunSetTracksOneAtomicUpdateAndKeepsPlainStreamsControlFree(t *testing.T) {
	var stdout, diagnostics bytes.Buffer
	var requests []SetRequest
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Capabilities: terminalexperience.Capabilities{Interaction: terminalexperience.PlainInteractive},
		Output:       &stdout,
		Diagnostics:  &diagnostics,
	})
	err := runSet(&Options{
		Context:  context.Background(),
		Profile:  "work",
		Key:      "model",
		Value:    "next-model",
		Terminal: experience,
		Store: func() (SetWriter, error) {
			return setWriterFunc(func(name, key, value string) error {
				requests = append(requests, SetRequest{Profile: name, Key: key, Value: value})
				return nil
			}), nil
		},
	})
	if err != nil {
		t.Fatalf("runSet() error = %v", err)
	}
	if got, want := requests, []SetRequest{{Profile: "work", Key: "model", Value: "next-model"}}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("writer requests = %#v, want %#v", got, want)
	}
	if got, want := stdout.String(), "Profile work updated\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
	if got, want := diagnostics.String(), "Updating CM profile...\n"; got != want {
		t.Fatalf("diagnostics = %q, want %q", got, want)
	}
	if terminaltest.ContainsTerminalControl(append(stdout.Bytes(), diagnostics.Bytes()...)) {
		t.Fatalf("streams contain terminal controls: stdout=%q diagnostics=%q", stdout.String(), diagnostics.String())
	}
}

func TestRunSetAutomationStaysSilentWhilePreservingTheSingleWriterCall(t *testing.T) {
	var stdout, diagnostics bytes.Buffer
	var calls int
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Capabilities: terminalexperience.Capabilities{Interaction: terminalexperience.Automation},
		Output:       &stdout,
		Diagnostics:  &diagnostics,
	})
	err := runSet(&Options{
		Context:  context.Background(),
		Profile:  "work",
		Key:      "apiKey",
		Value:    "secret-that-must-not-be-printed",
		Terminal: experience,
		Store: func() (SetWriter, error) {
			return setWriterFunc(func(_, _, _ string) error {
				calls++
				return nil
			}), nil
		},
	})
	if err != nil {
		t.Fatalf("runSet() error = %v", err)
	}
	if calls != 1 {
		t.Fatalf("writer calls = %d, want 1", calls)
	}
	if got, want := stdout.String(), "Profile work updated\n"; got != want {
		t.Fatalf("Automation result = %q, want %q", got, want)
	}
	if diagnostics.Len() != 0 {
		t.Fatalf("Automation emitted diagnostics: %q", diagnostics.String())
	}
}

func TestRunSetFailureDoesNotProjectTheRequestedValue(t *testing.T) {
	var stdout, diagnostics bytes.Buffer
	failure := errors.New("storage failed: secret-that-must-not-be-printed")
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Capabilities: terminalexperience.Capabilities{Interaction: terminalexperience.PlainInteractive},
		Output:       &stdout,
		Diagnostics:  &diagnostics,
	})
	err := runSet(&Options{
		Context:  context.Background(),
		Profile:  "work",
		Key:      "apiKey",
		Value:    "secret-that-must-not-be-printed",
		Terminal: experience,
		Store: func() (SetWriter, error) {
			return setWriterFunc(func(_, _, _ string) error { return failure }), nil
		},
	})
	if !errors.Is(err, failure) {
		t.Fatalf("runSet() error = %v, want %v", err, failure)
	}
	if stdout.Len() != 0 {
		t.Fatalf("failed update wrote stdout: %q", stdout.String())
	}
	if strings.Contains(diagnostics.String(), "secret-that-must-not-be-printed") {
		t.Fatalf("failed update leaked value in diagnostics: %q", diagnostics.String())
	}
}

func TestCMSetSuccessDetailUsesSafeKeySpecificProjections(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		contains []string
		absent   []string
	}{
		{name: "api key", key: "apiKey", value: "secret", contains: []string{"API key: [redacted]"}, absent: []string{"secret"}},
		{name: "url", key: "baseURL", value: "https://user:pass@example.test/v1?token=hidden#frag", contains: []string{"Base URL: https://example.test/v1"}, absent: []string{"user", "pass", "hidden", "frag"}},
		{name: "unsafe model", key: "model", value: "bad\nmodel", contains: []string{"Model configured"}, absent: []string{"bad\n"}},
		{name: "numeric", key: "timeoutMs", value: "1500suffix", contains: []string{"Requested value: 1500suffix"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := cmSetSuccessDetail("work", test.key, test.value)
			for _, expected := range test.contains {
				if !strings.Contains(got, expected) {
					t.Fatalf("detail = %q, missing %q", got, expected)
				}
			}
			for _, forbidden := range test.absent {
				if strings.Contains(got, forbidden) {
					t.Fatalf("detail = %q, leaked %q", got, forbidden)
				}
			}
		})
	}
}

type setWriterFunc func(name, key, value string) error

func (function setWriterFunc) SetCMProfileValue(name, key, value string) error {
	return function(name, key, value)
}

func TestTerminalCMSetPresentationPreservesPlainAndAutomationResults(t *testing.T) {
	result := SetResult{Profile: "work"}
	const want = "Profile work updated\n"

	for _, session := range []terminalexperience.Capabilities{
		{Interaction: terminalexperience.PlainInteractive},
		{Interaction: terminalexperience.Automation},
	} {
		var output bytes.Buffer
		experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{Capabilities: session, Output: &output})
		run := experience.Open(context.Background())
		if err := run.Result(terminalCMSetDocument(result)); err != nil {
			t.Fatalf("Present() error = %v", err)
		}
		if err := run.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
		if got := output.String(); got != want {
			t.Fatalf("%v result = %q, want %q", session.Interaction, got, want)
		}
		if terminaltest.ContainsTerminalControl(output.Bytes()) {
			t.Fatalf("%v result contains terminal control: %q", session.Interaction, output.String())
		}
	}
}

func TestTerminalCMSetPresentationUsesARichSuccessRole(t *testing.T) {
	for _, session := range []terminalexperience.Capabilities{
		{Interaction: terminalexperience.RichInteractive},
		{Interaction: terminalexperience.RichInteractive},
	} {
		document := terminalCMSetDocument(SetResult{Profile: "work"})
		if got, want := document.Blocks[0].Role, terminalexperience.VisualRoleSuccess; got != want {
			t.Fatalf("Rich role = %v, want %v", got, want)
		}
		var output bytes.Buffer
		experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{Capabilities: session, Output: &output})
		run := experience.Open(context.Background())
		if err := run.Result(document); err != nil {
			t.Fatalf("Present() error = %v", err)
		}
		if err := run.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
		if got, want := output.String(), "Profile work updated\n"; got != want {
			t.Fatalf("Rich output = %q, want %q", got, want)
		}
		if terminaltest.ContainsTerminalControl(output.Bytes()) {
			t.Fatalf("non-terminal writer output contains terminal control: %q", output.String())
		}
	}
}
