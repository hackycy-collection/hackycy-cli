package terminal_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"reflect"
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
		Capabilities: terminal.Capabilities{Interaction: terminal.PlainInteractive},
		Input:        strings.NewReader("project\n"),
		Output:       &stdout,
		Diagnostics:  &stderr,
	})
	run := experience.Open(context.Background())

	if err := run.Notice(terminal.PresentationDocument{Blocks: []terminal.PresentationBlock{{Text: "context"}}}); err != nil {
		t.Fatalf("Notice() error = %v", err)
	}
	answer, err := run.Ask(terminal.InteractionRequest{Kind: terminal.InteractionText, Message: "Project"})
	if err != nil || answer.Value != "project" {
		t.Fatalf("Ask() = (%#v, %v)", answer, err)
	}
	if _, err := io.WriteString(experience.DiagnosticWriter(), "diagnostic\n"); err != nil {
		t.Fatalf("diagnostic write error = %v", err)
	}
	if err := run.Result(terminal.PresentationDocument{Blocks: []terminal.PresentationBlock{{Text: "result"}}}); err != nil {
		t.Fatalf("Result() error = %v", err)
	}
	if err := run.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	if got, want := stdout.String(), "result\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
	if got, want := stderr.String(), "context\nProject: diagnostic\n"; got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
}

func TestPlainExperienceTrackDoesNotBufferDiagnostics(t *testing.T) {
	stderr := newPromptBuffer("Scanning")
	experience := terminal.NewExperience(terminal.ExperienceOptions{
		Capabilities: terminal.Capabilities{Interaction: terminal.PlainInteractive},
		Diagnostics:  stderr,
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
	if got, want := stderr.String(), "Scanning\ndeferred diagnostic\n"; got != want {
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

func TestPlainExperienceAskDoesNotBufferDiagnostics(t *testing.T) {
	input := newBlockingReader("project\n")
	stderr := newPromptBuffer("Project:")
	experience := terminal.NewExperience(terminal.ExperienceOptions{
		Capabilities: terminal.Capabilities{Interaction: terminal.PlainInteractive},
		Input:        input,
		Diagnostics:  stderr,
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
	if got, want := stderr.String(), "Project: deferred diagnostic\n"; got != want {
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
	experience := terminal.NewExperience(terminal.ExperienceOptions{Capabilities: terminal.Capabilities{Interaction: terminal.PlainInteractive}})
	run := experience.Open(context.Background())
	if err := run.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := run.Result(terminal.PresentationDocument{}); !errors.Is(err, terminal.ErrExperienceRunClosed) {
		t.Fatalf("Result() error = %v, want ErrExperienceRunClosed", err)
	}
	if _, err := run.Ask(terminal.InteractionRequest{Kind: terminal.InteractionText}); !errors.Is(err, terminal.ErrExperienceRunClosed) {
		t.Fatalf("Ask() error = %v, want ErrExperienceRunClosed", err)
	}
	if err := run.Track(terminal.TrackedOperation{}); !errors.Is(err, terminal.ErrExperienceRunClosed) {
		t.Fatalf("Track() error = %v, want ErrExperienceRunClosed", err)
	}
}

func TestExperienceFirstResultEndsInteractiveOperationsButAllowsMoreResults(t *testing.T) {
	var stdout bytes.Buffer
	experience := terminal.NewExperience(terminal.ExperienceOptions{
		Capabilities: terminal.Capabilities{Interaction: terminal.PlainInteractive},
		Output:       &stdout,
	})
	run := experience.Open(context.Background())
	if err := run.Result(terminal.PresentationDocument{Blocks: []terminal.PresentationBlock{{Text: "first"}}}); err != nil {
		t.Fatalf("first Result() error = %v", err)
	}
	if err := run.Result(terminal.PresentationDocument{Blocks: []terminal.PresentationBlock{{Text: "second"}}}); err != nil {
		t.Fatalf("second Result() error = %v", err)
	}
	if err := run.Notice(terminal.PresentationDocument{}); !errors.Is(err, terminal.ErrExperienceRunFinished) {
		t.Fatalf("Notice() error = %v, want ErrExperienceRunFinished", err)
	}
	if _, err := run.Ask(terminal.InteractionRequest{}); !errors.Is(err, terminal.ErrExperienceRunFinished) {
		t.Fatalf("Ask() error = %v, want ErrExperienceRunFinished", err)
	}
	if err := run.Track(terminal.TrackedOperation{}); !errors.Is(err, terminal.ErrExperienceRunFinished) {
		t.Fatalf("Track() error = %v, want ErrExperienceRunFinished", err)
	}
	if got, want := stdout.String(), "first\nsecond\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestRichPreflightSizeFailureFallsBackToPlainWithoutChangingCapabilities(t *testing.T) {
	input, err := os.CreateTemp(t.TempDir(), "input")
	if err != nil {
		t.Fatalf("create input: %v", err)
	}
	defer input.Close()
	if _, err := input.WriteString("project\n"); err != nil {
		t.Fatalf("write input: %v", err)
	}
	if _, err := input.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("rewind input: %v", err)
	}
	diagnostics, err := os.CreateTemp(t.TempDir(), "diagnostics")
	if err != nil {
		t.Fatalf("create diagnostics: %v", err)
	}
	defer diagnostics.Close()

	capabilities := terminal.Capabilities{
		Interaction: terminal.RichInteractive,
		Stdin:       terminal.StreamCapability{Terminal: true},
		Stderr:      terminal.StreamCapability{Terminal: true, Color: true},
	}
	experience := terminal.NewExperience(terminal.ExperienceOptions{
		Capabilities: capabilities,
		Input:        input,
		Diagnostics:  diagnostics,
	})
	run := experience.Open(context.Background())
	answer, err := run.Ask(terminal.InteractionRequest{Kind: terminal.InteractionText, Message: "Project"})
	if err != nil || answer.Value != "project" {
		t.Fatalf("Ask() = (%#v, %v)", answer, err)
	}
	if err := run.Notice(terminal.PresentationDocument{Blocks: []terminal.PresentationBlock{{Text: "plain notice"}}}); err != nil {
		t.Fatalf("Notice() error = %v", err)
	}
	if err := run.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	contents, err := os.ReadFile(diagnostics.Name())
	if err != nil {
		t.Fatalf("read diagnostics: %v", err)
	}
	if got, want := string(contents), "Project: plain notice\n"; got != want {
		t.Fatalf("diagnostics = %q, want %q", got, want)
	}
	if got := experience.Capabilities(); got != capabilities {
		t.Fatalf("Capabilities() = %#v, want %#v", got, capabilities)
	}
}

func TestResultWriteFailureStillEndsInteractionAndLaterResultsAreAttempted(t *testing.T) {
	first := errors.New("first result write")
	second := errors.New("second result write")
	output := &sequenceErrorWriter{errors: []error{first, second}}
	experience := terminal.NewExperience(terminal.ExperienceOptions{
		Capabilities: terminal.Capabilities{Interaction: terminal.PlainInteractive},
		Output:       output,
	})
	run := experience.Open(context.Background())
	if err := run.Result(terminal.PresentationDocument{Blocks: []terminal.PresentationBlock{{Text: "first"}}}); !errors.Is(err, first) {
		t.Fatalf("first Result() error = %v", err)
	}
	if err := run.Notice(terminal.PresentationDocument{}); !errors.Is(err, terminal.ErrExperienceRunFinished) {
		t.Fatalf("Notice() error = %v, want ErrExperienceRunFinished", err)
	}
	if err := run.Result(terminal.PresentationDocument{Blocks: []terminal.PresentationBlock{{Text: "second"}}}); !errors.Is(err, second) {
		t.Fatalf("second Result() error = %v", err)
	}
	if got, want := output.writes, []string{"first\n", "second\n"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("writes = %#v, want %#v", got, want)
	}
	if err := run.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
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
