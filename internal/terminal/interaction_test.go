package terminal_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestAutomationAskFailsBeforeReadingOrWriting(t *testing.T) {
	handler := terminal.NewInteractionHandler(terminal.InteractionOptions{
		Session:     terminal.Session{Kind: terminal.Automation},
		Input:       panicReader{},
		Diagnostics: panicWriter{},
	})

	_, err := handler.Ask(context.Background(), terminal.InteractionRequest{Kind: terminal.InteractionText, Message: "Name"})
	if !errors.Is(err, terminal.ErrAutomationInteraction) {
		t.Fatalf("Ask() error = %v, want ErrAutomationInteraction", err)
	}
}

func TestPlainAskUsesDiagnosticStreamAndRetriesValidation(t *testing.T) {
	var diagnostics bytes.Buffer
	handler := terminal.NewInteractionHandler(terminal.InteractionOptions{
		Session:     terminal.Session{Kind: terminal.PlainInteractive},
		Input:       strings.NewReader("\nproject\n"),
		Diagnostics: &diagnostics,
	})

	answer, err := handler.Ask(context.Background(), terminal.InteractionRequest{
		Kind:        terminal.InteractionText,
		Message:     "Project name",
		Description: "Choose a local project.",
		Placeholder: "example",
		Validate: func(answer terminal.InteractionAnswer) error {
			if answer.Value == "" {
				return errors.New("project name is required")
			}
			return nil
		},
	})
	if err != nil || answer.Value != "project" {
		t.Fatalf("Ask() = (%#v, %v)", answer, err)
	}
	if got := diagnostics.String(); !strings.Contains(got, "Choose a local project.") || !strings.Contains(got, "project name is required") {
		t.Fatalf("diagnostics = %q", got)
	}
	if terminaltest.ContainsTerminalControl(diagnostics.Bytes()) {
		t.Fatalf("plain diagnostics contain terminal control: %q", diagnostics.String())
	}
}

func TestPlainAskSelectsExistingOptionsWithoutImplicitDefault(t *testing.T) {
	var diagnostics bytes.Buffer
	handler := terminal.NewInteractionHandler(terminal.InteractionOptions{
		Session:     terminal.Session{Kind: terminal.PlainInteractive},
		Input:       strings.NewReader("wrong\n2\n"),
		Diagnostics: &diagnostics,
	})

	answer, err := handler.Ask(context.Background(), terminal.InteractionRequest{
		Kind:    terminal.InteractionSelect,
		Message: "Choose an environment",
		Options: []terminal.InteractionOption{
			{Label: "Development", Value: "dev"},
			{Label: "Production", Value: "prod"},
		},
	})
	if err != nil || answer.Value != "prod" {
		t.Fatalf("Ask() = (%#v, %v)", answer, err)
	}
	if !strings.Contains(diagnostics.String(), "invalid selection") {
		t.Fatalf("diagnostics = %q, want invalid-selection feedback", diagnostics.String())
	}
}

func TestRichAskUsesExplicitPTYInputAndDiagnosticOutput(t *testing.T) {
	const helperEnvironment = "YCY_TERMINAL_HUH_HELPER"
	if os.Getenv(helperEnvironment) == "1" {
		handler := terminal.NewInteractionHandler(terminal.InteractionOptions{
			Session:     terminal.Session{Kind: terminal.RichInteractive, Color: true},
			Input:       os.Stdin,
			Diagnostics: os.Stderr,
		})
		answer, err := handler.Ask(context.Background(), terminal.InteractionRequest{
			Kind:    terminal.InteractionText,
			Message: "Project name",
			Validate: func(answer terminal.InteractionAnswer) error {
				if answer.Value == "" {
					return errors.New("project name is required")
				}
				return nil
			},
		})
		if err != nil {
			t.Fatalf("Ask() error = %v", err)
		}
		_, _ = fmt.Fprintf(os.Stdout, "answer=%s\n", answer.Value)
		return
	}

	command := exec.Command(os.Args[0], "-test.run=^TestRichAskUsesExplicitPTYInputAndDiagnosticOutput$")
	command.Env = append(richPTYEnvironment(), helperEnvironment+"=1", "TERM=xterm-256color")
	process, err := terminaltest.StartPTY(command)
	if errors.Is(err, terminaltest.ErrPTYUnsupported) {
		t.Skip(err)
	}
	if err != nil {
		t.Fatalf("start PTY helper: %v", err)
	}
	defer process.Close()

	output := newPromptBuffer("submit")
	readDone := make(chan struct{})
	go func() {
		_, _ = io.Copy(output, process.Terminal())
		close(readDone)
	}()
	respondToHuhTerminalQueries(t, process, output)
	select {
	case <-output.prompt:
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for Huh prompt: %q", output.String())
	}
	if _, err := io.WriteString(process.Terminal(), "ycy\r"); err != nil {
		t.Fatalf("write PTY input: %v", err)
	}
	if err := process.Wait(); err != nil {
		t.Fatalf("wait PTY helper: %v", err)
	}
	if err := process.Close(); err != nil {
		t.Fatalf("close PTY helper: %v", err)
	}
	select {
	case <-readDone:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out reading PTY output")
	}
	assertRichCleanup(t, output.String(), "answer=ycy")
}

func TestRichAskCtrlCRestoresTerminalBeforeReturningCancellation(t *testing.T) {
	const helperEnvironment = "YCY_TERMINAL_HUH_CANCEL_HELPER"
	if os.Getenv(helperEnvironment) == "1" {
		handler := terminal.NewInteractionHandler(terminal.InteractionOptions{
			Session:     terminal.Session{Kind: terminal.RichInteractive, Color: true},
			Input:       os.Stdin,
			Diagnostics: os.Stderr,
		})
		_, err := handler.Ask(context.Background(), terminal.InteractionRequest{Kind: terminal.InteractionText, Message: "Project name"})
		if !errors.Is(err, terminal.ErrInteractionCancelled) {
			t.Fatalf("Ask() error = %v, want ErrInteractionCancelled", err)
		}
		_, _ = fmt.Fprintln(os.Stdout, "cancelled=true")
		return
	}

	command := exec.Command(os.Args[0], "-test.run=^TestRichAskCtrlCRestoresTerminalBeforeReturningCancellation$")
	command.Env = append(richPTYEnvironment(), helperEnvironment+"=1", "TERM=xterm-256color")
	process, err := terminaltest.StartPTY(command)
	if errors.Is(err, terminaltest.ErrPTYUnsupported) {
		t.Skip(err)
	}
	if err != nil {
		t.Fatalf("start PTY helper: %v", err)
	}
	defer process.Close()

	output := newPromptBuffer("submit")
	readDone := make(chan struct{})
	go func() {
		_, _ = io.Copy(output, process.Terminal())
		close(readDone)
	}()
	respondToHuhTerminalQueries(t, process, output)
	select {
	case <-output.prompt:
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for Huh prompt: %q", output.String())
	}
	if _, err := io.WriteString(process.Terminal(), "\x03"); err != nil {
		t.Fatalf("write PTY Ctrl-C: %v", err)
	}
	if err := process.Wait(); err != nil {
		t.Fatalf("wait PTY helper: %v", err)
	}
	if err := process.Close(); err != nil {
		t.Fatalf("close PTY helper: %v", err)
	}
	select {
	case <-readDone:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out reading PTY output")
	}
	assertRichCleanup(t, output.String(), "cancelled=true")
}

type promptBuffer struct {
	mu        sync.Mutex
	buffer    bytes.Buffer
	needle    string
	prompt    chan struct{}
	query     chan struct{}
	once      sync.Once
	queryOnce sync.Once
}

func newPromptBuffer(needle string) *promptBuffer {
	return &promptBuffer{needle: needle, prompt: make(chan struct{}), query: make(chan struct{})}
}

func respondToHuhTerminalQueries(t *testing.T, process *terminaltest.PTYProcess, output *promptBuffer) {
	t.Helper()
	select {
	case <-output.query:
		if _, err := io.WriteString(process.Terminal(), "\x1b]11;rgb:0000/0000/0000\x1b\\\x1b[1;1R"); err != nil {
			t.Fatalf("write PTY terminal responses: %v", err)
		}
	case <-output.prompt:
		// The terminal may have enough cached capability state to render immediately.
	}
}

func assertRichCleanup(t *testing.T, output, marker string) {
	t.Helper()
	if !strings.Contains(output, marker) {
		t.Fatalf("PTY output missing %q: %q", marker, output)
	}
	if !strings.Contains(output, "\x1b[?25h") || !strings.Contains(output, "\x1b[?2004l") {
		t.Fatalf("Rich cleanup bytes missing from PTY output: %q", output)
	}
}

func (buffer *promptBuffer) Write(value []byte) (int, error) {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	count, err := buffer.buffer.Write(value)
	if strings.Contains(buffer.buffer.String(), buffer.needle) {
		buffer.once.Do(func() { close(buffer.prompt) })
	}
	if strings.Contains(buffer.buffer.String(), "\x1b]11;?") || strings.Contains(buffer.buffer.String(), "\x1b[6n") {
		buffer.queryOnce.Do(func() { close(buffer.query) })
	}
	return count, err
}

func (buffer *promptBuffer) String() string {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return buffer.buffer.String()
}

type panicReader struct{}

func (panicReader) Read([]byte) (int, error) {
	panic("Automation Session attempted to read implicit input")
}

type panicWriter struct{}

func (panicWriter) Write([]byte) (int, error) {
	panic("Automation Session attempted to write an interaction")
}
