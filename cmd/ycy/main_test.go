package main

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
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

func TestRootDiagnosticsDeferNormalWritesUntilRendererLeaseCloses(t *testing.T) {
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
	runtime := newRootLoggingRuntime(root)
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
