package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hackycy/hackycy-cli/internal/logging"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
	commandfactory "github.com/hackycy/hackycy-cli/pkg/cmd/factory"
	"github.com/hackycy/hackycy-cli/pkg/cmdutil"
)

func TestNewRootTerminalConstructsOneExperienceFromInheritedFacts(t *testing.T) {
	input, err := os.CreateTemp(t.TempDir(), "stdin")
	if err != nil {
		t.Fatalf("create stdin: %v", err)
	}
	defer input.Close()
	output, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatalf("create stdout: %v", err)
	}
	defer output.Close()
	diagnostics, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatalf("create stderr: %v", err)
	}
	defer diagnostics.Close()

	root := newRootTerminal(input, output, diagnostics, func(key string) (string, bool) {
		return map[string]string{"TERM": "xterm-256color"}[key], key == "TERM"
	}, func(*os.File) bool { return true })

	if root.input != input || root.output != output || root.diagnostics != diagnostics {
		t.Fatal("root terminal did not preserve the inherited stream identities")
	}
	if got, want := root.experience.Session(), (terminalexperience.Session{Kind: terminalexperience.RichInteractive, Color: true}); got != want {
		t.Fatalf("experience session = %#v, want %#v", got, want)
	}
}

func TestRootTunnelServerDiagnosticsPreserveSessionStreamContracts(t *testing.T) {
	for _, testCase := range []struct {
		name        string
		environment map[string]string
		terminal    bool
		format      logging.RecordFormat
		styled      bool
	}{
		{name: "rich text", environment: map[string]string{"TERM": "xterm-256color"}, terminal: true, styled: true},
		{name: "no color text", environment: map[string]string{"TERM": "xterm-256color", "NO_COLOR": "1"}, terminal: true},
		{name: "redirected JSON", environment: map[string]string{"TERM": "xterm-256color"}, format: logging.JSONFormat},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input, err := os.CreateTemp(t.TempDir(), "stdin")
			if err != nil {
				t.Fatalf("create stdin: %v", err)
			}
			defer input.Close()
			output, err := os.CreateTemp(t.TempDir(), "stdout")
			if err != nil {
				t.Fatalf("create stdout: %v", err)
			}
			defer output.Close()
			diagnostics, err := os.CreateTemp(t.TempDir(), "stderr")
			if err != nil {
				t.Fatalf("create stderr: %v", err)
			}
			defer diagnostics.Close()

			root := newRootTerminal(input, output, diagnostics, func(key string) (string, bool) {
				value, ok := testCase.environment[key]
				return value, ok
			}, func(*os.File) bool { return testCase.terminal })
			factory := newCommandFactoryForRootTerminal(root)
			root.experience = factory.Terminal
			runtime := factory.Logging
			runtime.SetFormat(testCase.format)
			runtime.Logger("tunnel.server").Info("Tunnel started Bearer server-secret", map[string]any{
				"authorization": "authorization-secret",
				"port":          7000,
			})

			stdout := readRootDiagnostic(t, output)
			stderr := readRootDiagnostic(t, diagnostics)
			if stdout != "" {
				t.Fatalf("stdout = %q, want empty", stdout)
			}
			for _, secret := range []string{"server-secret", "authorization-secret"} {
				if strings.Contains(stderr, secret) {
					t.Fatalf("stderr leaked %q: %q", secret, stderr)
				}
			}
			if !strings.Contains(stderr, "[REDACTED]") {
				t.Fatalf("stderr omitted redaction marker: %q", stderr)
			}
			if testCase.format == logging.JSONFormat {
				var record map[string]any
				if err := json.Unmarshal([]byte(stderr), &record); err != nil {
					t.Fatalf("stderr is not NDJSON: %v; output = %q", err, stderr)
				}
				if record["scope"] != "tunnel.server" || record["level"] != "info" {
					t.Fatalf("NDJSON record = %#v", record)
				}
			}
			if got := terminaltest.ContainsTerminalControl([]byte(stderr)); got != testCase.styled {
				t.Fatalf("terminal control = %t, want %t; stderr = %q", got, testCase.styled, stderr)
			}
		})
	}
}

func TestRootTunnelServerDiagnosticsDeferUntilRendererLeaseCloses(t *testing.T) {
	input, err := os.CreateTemp(t.TempDir(), "stdin")
	if err != nil {
		t.Fatalf("create stdin: %v", err)
	}
	defer input.Close()
	output, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatalf("create stdout: %v", err)
	}
	defer output.Close()
	diagnostics, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatalf("create stderr: %v", err)
	}
	defer diagnostics.Close()

	root := newRootTerminal(input, output, diagnostics, func(key string) (string, bool) {
		return map[string]string{"TERM": "dumb"}[key], key == "TERM"
	}, func(*os.File) bool { return true })
	factory := newCommandFactoryForRootTerminal(root)
	root.experience = factory.Terminal
	runtime := factory.Logging
	run := root.experience.Open(context.Background())
	updates := make(chan terminalexperience.OperationPhase)
	tracked := make(chan error, 1)
	updatesClosed := false
	trackFinished := false
	defer func() {
		if !updatesClosed {
			close(updates)
		}
		if !trackFinished {
			<-tracked
		}
		_ = run.Close()
	}()

	go func() {
		tracked <- run.Track(terminalexperience.TrackedOperation{Updates: updates})
	}()
	updates <- terminalexperience.OperationPhase{Name: "Scanning", State: terminalexperience.PhaseActive}
	waitForRootDiagnostic(t, diagnostics, "Scanning\n")

	if _, err := io.WriteString(root.experience.DiagnosticWriter(), "cobra diagnostic\n"); err != nil {
		t.Fatalf("write normal diagnostic: %v", err)
	}
	runtime.Logger("tunnel.server").Info("logger diagnostic", nil)
	if got := readRootDiagnostic(t, diagnostics); strings.Contains(got, "cobra diagnostic") || strings.Contains(got, "logger diagnostic") {
		t.Fatalf("normal diagnostics wrote during an active lease: %q", got)
	}

	close(updates)
	updatesClosed = true
	select {
	case err := <-tracked:
		trackFinished = true
		if err != nil {
			t.Fatalf("Track() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Track() did not release the renderer lease")
	}

	got := readRootDiagnostic(t, diagnostics)
	for _, text := range []string{"Scanning\n", "cobra diagnostic\n", "logger diagnostic"} {
		if !strings.Contains(got, text) {
			t.Fatalf("diagnostics = %q, missing %q", got, text)
		}
	}
	if strings.Index(got, "Scanning\n") > strings.Index(got, "cobra diagnostic\n") || strings.Index(got, "cobra diagnostic\n") > strings.Index(got, "logger diagnostic") {
		t.Fatalf("diagnostics did not flush after renderer output in write order: %q", got)
	}
}

func waitForRootDiagnostic(t *testing.T, file *os.File, expected string) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if got := readRootDiagnostic(t, file); strings.Contains(got, expected) {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("diagnostics did not contain %q before timeout: %q", expected, readRootDiagnostic(t, file))
}

func readRootDiagnostic(t *testing.T, file *os.File) string {
	t.Helper()
	contents, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatalf("read diagnostics: %v", err)
	}
	return string(contents)
}

func newCommandFactoryForRootTerminal(root rootTerminal) *cmdutil.Factory {
	return commandfactory.New(commandfactory.Options{
		Version: "0.0.0-dev",
		IOStreams: cmdutil.IOStreams{
			In:     root.input,
			Out:    root.output,
			ErrOut: root.diagnostics,
		},
		Session: root.experience.Session(),
	})
}
