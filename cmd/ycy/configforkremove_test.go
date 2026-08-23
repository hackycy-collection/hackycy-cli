package main

import (
	"bytes"
	"strings"
	"testing"

	configfork "github.com/hackycy/hackycy-cli/internal/commands/config/fork"
)

func TestTerminalForkRemovePrompterConfirmsAndDefaultsToNo(t *testing.T) {
	output := &bytes.Buffer{}
	prompter := newTerminalForkRemovePrompter(strings.NewReader("yes\n"), output)

	confirmed, cancelled := prompter.Confirm(configfork.ConfirmPrompt{Message: `Remove instance "work"?`})

	if !confirmed || cancelled {
		t.Fatalf("Confirm() = (%t, %t), want confirmation", confirmed, cancelled)
	}
	if !strings.Contains(output.String(), `Remove instance "work"? [y/N]:`) {
		t.Fatalf("confirmation output = %q", output.String())
	}

	defaultPrompter := newTerminalForkRemovePrompter(strings.NewReader("\n"), &bytes.Buffer{})
	confirmed, cancelled = defaultPrompter.Confirm(configfork.ConfirmPrompt{Message: "Remove instance \"work\"?"})
	if confirmed || cancelled {
		t.Fatalf("default Confirm() = (%t, %t), want negative confirmation", confirmed, cancelled)
	}
}

func TestTerminalForkRemovePrompterTreatsEOFAsCancellation(t *testing.T) {
	prompter := newTerminalForkRemovePrompter(strings.NewReader(""), &bytes.Buffer{})

	confirmed, cancelled := prompter.Confirm(configfork.ConfirmPrompt{Message: "Remove instance \"work\"?"})

	if confirmed || !cancelled {
		t.Fatalf("Confirm() = (%t, %t), want cancellation", confirmed, cancelled)
	}
}

func TestTerminalForkRemovePresenterWritesOutcomeMessages(t *testing.T) {
	output := &bytes.Buffer{}
	presenter := terminalForkRemovePresenter{output: output}

	presenter.Info("No instances configured")
	presenter.Outcome("Nothing to remove")
	presenter.Outcome("Cancelled")
	presenter.Outcome("Instance work removed")

	if got, want := output.String(), "No instances configured\nNothing to remove\nCancelled\nInstance work removed\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}
