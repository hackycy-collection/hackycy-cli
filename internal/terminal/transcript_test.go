package terminal

import (
	"strings"
	"testing"
)

func TestTranscriptLedgerNormalizesAndNumbersEvents(t *testing.T) {
	ledger := NewTranscriptLedger(TranscriptOptions{MaxEvents: 8, MaxBytes: 4096, MaxFieldSize: 64})
	ledger.Append(TranscriptEvent{Kind: TranscriptAsk, Label: " Project\tname ", Text: "value\r\nnext\x1b[2K"})
	ledger.Append(TranscriptEvent{Kind: TranscriptMilestone, Text: "checkpoint"})

	events := ledger.Events()
	if len(events) != 2 || events[0].Sequence != 1 || events[1].Sequence != 2 {
		t.Fatalf("events = %#v", events)
	}
	if events[0].Label != "Project name" || events[0].Text != "value next" {
		t.Fatalf("normalized event = %#v", events[0])
	}
	if got, want := ledger.Render(), "Project name: value next\ncheckpoint\n"; got != want {
		t.Fatalf("render = %q, want %q", got, want)
	}
}

func TestTranscriptLedgerBoundsAtEventBoundariesAndMarksTruncation(t *testing.T) {
	ledger := NewTranscriptLedger(TranscriptOptions{MaxEvents: 4, MaxBytes: 256, MaxFieldSize: 32})
	for index := 0; index < 20; index++ {
		ledger.Append(TranscriptEvent{Kind: TranscriptMilestone, Text: strings.Repeat("x", 20)})
	}
	events := ledger.Events()
	if len(events) > 4 {
		t.Fatalf("event count = %d, want <= 4", len(events))
	}
	if len(events) == 0 || events[len(events)-1].Kind != TranscriptTruncated {
		t.Fatalf("events missing truncation marker: %#v", events)
	}
	if got := strings.Count(ledger.Render(), "... transcript truncated ..."); got != 1 {
		t.Fatalf("truncation marker count = %d", got)
	}
	before := len(events)
	ledger.Append(TranscriptEvent{Kind: TranscriptMilestone, Text: "late"})
	if got := len(ledger.Events()); got != before {
		t.Fatalf("append after truncation changed event count from %d to %d", before, got)
	}
}

func TestTranscriptLedgerRetainsFinalOutcomeAfterTruncation(t *testing.T) {
	ledger := NewTranscriptLedger(TranscriptOptions{MaxEvents: 4, MaxBytes: 512, MaxFieldSize: 64})
	for index := 0; index < 20; index++ {
		ledger.Append(TranscriptEvent{Kind: TranscriptMilestone, Text: "checkpoint"})
	}
	ledger.Append(TranscriptEvent{Kind: TranscriptOutcome, Outcome: Succeeded})

	events := ledger.Events()
	if len(events) > 4 {
		t.Fatalf("event count = %d, want <= 4", len(events))
	}
	if len(events) < 2 || events[len(events)-2].Kind != TranscriptTruncated || events[len(events)-1].Kind != TranscriptOutcome {
		t.Fatalf("events = %#v, want truncation followed by outcome", events)
	}
	for index, event := range events {
		if event.Sequence != uint64(index+1) {
			t.Fatalf("event sequence at %d = %d, want %d", index, event.Sequence, index+1)
		}
	}
	if got := ledger.Render(); !strings.HasSuffix(got, "succeeded\n") {
		t.Fatalf("render = %q, want final outcome", got)
	}
}

func TestTranscriptLedgerNormalizesTinyBudgetsAroundMandatoryEvents(t *testing.T) {
	ledger := NewTranscriptLedger(TranscriptOptions{MaxEvents: 1, MaxBytes: 1, MaxFieldSize: 1})
	ledger.Append(TranscriptEvent{Kind: TranscriptMilestone, Text: "discarded"})
	ledger.Append(TranscriptEvent{Kind: TranscriptOutcome, Outcome: Failed})

	events := ledger.Events()
	if len(events) != 2 || events[0].Kind != TranscriptTruncated || events[1].Kind != TranscriptOutcome {
		t.Fatalf("events = %#v, want truncation followed by outcome", events)
	}
	if events[0].Text != transcriptTruncatedText {
		t.Fatalf("truncation marker = %q, want %q", events[0].Text, transcriptTruncatedText)
	}
	if events[0].Sequence != 1 || events[1].Sequence != 2 {
		t.Fatalf("sequences = %d, %d, want 1, 2", events[0].Sequence, events[1].Sequence)
	}
	if got := ledger.Render(); got != "... transcript truncated ...\nfailed\n" {
		t.Fatalf("render = %q, want marker and final outcome", got)
	}
}

func TestTranscriptLedgerIgnoresEventsAfterOutcome(t *testing.T) {
	ledger := NewTranscriptLedger(TranscriptOptions{})
	ledger.Append(TranscriptEvent{Kind: TranscriptOutcome, Outcome: Succeeded})
	ledger.Append(TranscriptEvent{Kind: TranscriptMilestone, Text: "late"})
	ledger.Append(TranscriptEvent{Kind: TranscriptOutcome, Outcome: Failed})

	events := ledger.Events()
	if len(events) != 1 || events[0].Outcome != Succeeded {
		t.Fatalf("events after outcome changed ledger: %#v", events)
	}
}

func TestTranscriptLedgerFreezesAndReturnsCopies(t *testing.T) {
	ledger := NewTranscriptLedger(TranscriptOptions{})
	ledger.Append(TranscriptEvent{Kind: TranscriptMilestone, Text: "first"})
	events := ledger.Freeze()
	events[0].Text = "mutated"
	ledger.Append(TranscriptEvent{Kind: TranscriptMilestone, Text: "second"})
	got := ledger.Events()
	if len(got) != 1 || got[0].Text != "first" {
		t.Fatalf("frozen events = %#v", got)
	}
}

func TestInteractionTranscriptProjectionUsesSafeSemanticValues(t *testing.T) {
	selectRequest := InteractionRequest{
		Kind:    InteractionSelect,
		Options: []InteractionOption{{Label: "Production", Value: "prod"}},
	}
	if got := interactionTranscriptText(selectRequest, InteractionAnswer{Value: "prod"}); got != "Production" {
		t.Fatalf("select transcript = %q, want label", got)
	}
	multiRequest := InteractionRequest{
		Kind:    InteractionMultiSelect,
		Options: []InteractionOption{{Label: "One", Value: "one"}, {Label: "Two", Value: "two"}},
	}
	if got := interactionTranscriptText(multiRequest, InteractionAnswer{Values: []string{"two", "one"}}); got != "Two, One" {
		t.Fatalf("multi transcript = %q, want ordered labels", got)
	}
	for _, request := range []InteractionRequest{
		{Kind: InteractionSecret},
		{Kind: InteractionText, Sensitive: true},
	} {
		if got := interactionTranscriptText(request, InteractionAnswer{Value: "secret-value"}); got != "[redacted]" {
			t.Fatalf("sensitive transcript = %q, want redaction", got)
		}
	}
	if got := (PresentationDocument{Blocks: []PresentationBlock{{Text: "token", Sensitive: true}, {Text: "safe"}}}).transcriptText(); got != "[redacted] safe" {
		t.Fatalf("document transcript = %q", got)
	}
}
