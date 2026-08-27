package terminal_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hackycy/hackycy-cli/internal/terminal"
)

func TestExperienceRoutesPresentationAndInteractionToSeparateStreams(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	experience := terminal.NewExperience(terminal.ExperienceOptions{
		Session:     terminal.Session{Kind: terminal.PlainInteractive},
		Input:       strings.NewReader("project\n"),
		Output:      &stdout,
		Diagnostics: &stderr,
	})
	run := experience.Open(context.Background())

	if err := run.Present(terminal.PresentationDocument{Blocks: []terminal.PresentationBlock{{Text: "result"}}}); err != nil {
		t.Fatalf("Present() error = %v", err)
	}
	answer, err := run.Ask(terminal.InteractionRequest{Kind: terminal.InteractionText, Message: "Project"})
	if err != nil || answer.Value != "project" {
		t.Fatalf("Ask() = (%#v, %v)", answer, err)
	}
	if _, err := io.WriteString(experience.DiagnosticWriter(), "diagnostic\n"); err != nil {
		t.Fatalf("diagnostic write error = %v", err)
	}
	if err := run.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	if got, want := stdout.String(), "result\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
	if got, want := stderr.String(), "Project: diagnostic\n"; got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
}

func TestExperienceTrackDefersDiagnosticsUntilTheLeaseCloses(t *testing.T) {
	stderr := newPromptBuffer("Scanning")
	experience := terminal.NewExperience(terminal.ExperienceOptions{
		Session:     terminal.Session{Kind: terminal.PlainInteractive},
		Diagnostics: stderr,
	})
	run := experience.Open(context.Background())
	updates := make(chan terminal.OperationPhase)
	trackDone := make(chan error, 1)
	go func() {
		trackDone <- run.Track(terminal.TrackedOperation{Updates: updates})
	}()
	updates <- terminal.OperationPhase{Name: "Scanning", State: terminal.PhaseActive}
	select {
	case <-stderr.prompt:
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for tracked phase: %q", stderr.String())
	}
	if _, err := io.WriteString(experience.DiagnosticWriter(), "deferred diagnostic\n"); err != nil {
		t.Fatalf("diagnostic write error = %v", err)
	}
	if got, want := stderr.String(), "Scanning\n"; got != want {
		t.Fatalf("stderr during track = %q, want %q", got, want)
	}
	close(updates)
	select {
	case err := <-trackDone:
		if err != nil {
			t.Fatalf("Track() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Track() did not return after updates closed")
	}
	if got, want := stderr.String(), "Scanning\ndeferred diagnostic\n"; got != want {
		t.Fatalf("stderr after track = %q, want %q", got, want)
	}
}

func TestExperienceAskDefersDiagnosticsUntilTheLeaseCloses(t *testing.T) {
	input := newBlockingReader("project\n")
	stderr := newPromptBuffer("Project:")
	experience := terminal.NewExperience(terminal.ExperienceOptions{
		Session:     terminal.Session{Kind: terminal.PlainInteractive},
		Input:       input,
		Diagnostics: stderr,
	})
	run := experience.Open(context.Background())
	answerDone := make(chan struct {
		answer terminal.InteractionAnswer
		err    error
	}, 1)
	go func() {
		answer, err := run.Ask(terminal.InteractionRequest{Kind: terminal.InteractionText, Message: "Project"})
		answerDone <- struct {
			answer terminal.InteractionAnswer
			err    error
		}{answer: answer, err: err}
	}()
	select {
	case <-input.started:
	case <-time.After(time.Second):
		t.Fatal("Ask() did not begin reading input")
	}
	if _, err := io.WriteString(experience.DiagnosticWriter(), "deferred diagnostic\n"); err != nil {
		t.Fatalf("diagnostic write error = %v", err)
	}
	if got, want := stderr.String(), "Project: "; got != want {
		t.Fatalf("stderr during ask = %q, want %q", got, want)
	}
	close(input.release)
	select {
	case result := <-answerDone:
		if result.err != nil || result.answer.Value != "project" {
			t.Fatalf("Ask() = (%#v, %v)", result.answer, result.err)
		}
	case <-time.After(time.Second):
		t.Fatal("Ask() did not return after input became available")
	}
	if got, want := stderr.String(), "Project: deferred diagnostic\n"; got != want {
		t.Fatalf("stderr after ask = %q, want %q", got, want)
	}
}

func TestExperienceRunRejectsUseAfterClose(t *testing.T) {
	experience := terminal.NewExperience(terminal.ExperienceOptions{Session: terminal.Session{Kind: terminal.PlainInteractive}})
	run := experience.Open(context.Background())
	if err := run.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := run.Present(terminal.PresentationDocument{}); !errors.Is(err, terminal.ErrExperienceRunClosed) {
		t.Fatalf("Present() error = %v, want ErrExperienceRunClosed", err)
	}
	if _, err := run.Ask(terminal.InteractionRequest{Kind: terminal.InteractionText}); !errors.Is(err, terminal.ErrExperienceRunClosed) {
		t.Fatalf("Ask() error = %v, want ErrExperienceRunClosed", err)
	}
	if err := run.Track(terminal.TrackedOperation{}); !errors.Is(err, terminal.ErrExperienceRunClosed) {
		t.Fatalf("Track() error = %v, want ErrExperienceRunClosed", err)
	}
}

type blockingReader struct {
	payload string
	started chan struct{}
	release chan struct{}
	once    sync.Once
	read    bool
}

func newBlockingReader(payload string) *blockingReader {
	return &blockingReader{
		payload: payload,
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
}

func (reader *blockingReader) Read(buffer []byte) (int, error) {
	if reader.read {
		return 0, io.EOF
	}
	reader.once.Do(func() { close(reader.started) })
	<-reader.release
	reader.read = true
	return copy(buffer, reader.payload), nil
}
