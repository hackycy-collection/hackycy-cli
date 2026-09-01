package terminal

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestRuntimeRecoversStoppedRichRendererAndBlocksReplay(t *testing.T) {
	rendererErr := errors.New("renderer failed")
	var diagnostics bytes.Buffer
	runtime := NewExperience(ExperienceOptions{
		Capabilities: Capabilities{Interaction: RichInteractive},
		Diagnostics:  &diagnostics,
		Input:        strings.NewReader("unexpected input\n"),
	})
	run := runtime.Open(context.Background()).(*runtimeRun)
	run.ledger.Append(TranscriptEvent{Kind: TranscriptMilestone, Text: "safe checkpoint"})

	controller := &richController{
		runtime: runtime,
		done:    make(chan struct{}),
		err:     rendererErr,
		lease:   runtime.diagnostics.AcquireRendererLease(),
	}
	close(controller.done)
	run.controller = controller

	err := run.Notice(PresentationDocument{Blocks: []PresentationBlock{{Text: "unused"}}})
	if !errors.Is(err, rendererErr) {
		t.Fatalf("Notice() error = %v, want renderer failure", err)
	}
	if got, want := diagnostics.String(), "safe checkpoint\n"; got != want {
		t.Fatalf("replayed diagnostics = %q, want %q", got, want)
	}
	if run.controller != nil || !run.richDisabled || run.richFailure == nil {
		t.Fatalf("renderer recovery state = controller=%v disabled=%v failure=%v", run.controller, run.richDisabled, run.richFailure)
	}
	if _, err := io.WriteString(runtime.DiagnosticWriter(), "after recovery\n"); err != nil {
		t.Fatalf("diagnostic after recovery = %v", err)
	}
	if got, want := diagnostics.String(), "safe checkpoint\nafter recovery\n"; got != want {
		t.Fatalf("diagnostics after lease release = %q, want %q", got, want)
	}

	if _, err := run.Ask(InteractionRequest{Kind: InteractionText, Message: "Retry?"}); !errors.Is(err, rendererErr) {
		t.Fatalf("Ask() after renderer failure = %v, want original failure", err)
	}
	if err := run.Finish(Succeeded, &PresentationDocument{Blocks: []PresentationBlock{{Text: "result"}}}); !errors.Is(err, rendererErr) {
		t.Fatalf("Finish() after renderer failure = %v, want original failure", err)
	}
	if got := diagnostics.String(); strings.Contains(got, "unused") || strings.Contains(got, "result") {
		t.Fatalf("recovery emitted repeated semantic work: %q", got)
	}
	if err := run.Close(); err != nil {
		t.Fatalf("Close() after recovery = %v", err)
	}
}
