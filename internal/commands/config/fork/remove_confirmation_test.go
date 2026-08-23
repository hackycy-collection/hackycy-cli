package fork

import "testing"

func TestConfirmRemoveUsesLegacyQuestionAndAcceptsConfirmation(t *testing.T) {
	prompter := &scriptedRemoveConfirmationPrompter{confirmed: true}

	outcome := ConfirmRemove("work", prompter)

	if outcome != RemoveConfirmed {
		t.Fatalf("ConfirmRemove() = %v, want confirmed", outcome)
	}
	if got, want := prompter.question, (ConfirmPrompt{Message: `Remove instance "work"?`}); got != want {
		t.Fatalf("confirmation question = %#v, want %#v", got, want)
	}
}

func TestConfirmRemoveDoesNotEscapeAConfiguredAlias(t *testing.T) {
	prompter := &scriptedRemoveConfirmationPrompter{confirmed: true}

	ConfirmRemove(`work"alias`, prompter)

	if got, want := prompter.question, (ConfirmPrompt{Message: `Remove instance "work"alias"?`}); got != want {
		t.Fatalf("confirmation question = %#v, want %#v", got, want)
	}
}

func TestConfirmRemoveRetainsNegativeConfirmation(t *testing.T) {
	prompter := &scriptedRemoveConfirmationPrompter{}

	if outcome := ConfirmRemove("work", prompter); outcome != RemoveDeclined {
		t.Fatalf("ConfirmRemove() = %v, want declined", outcome)
	}
}

func TestConfirmRemoveRetainsCancellation(t *testing.T) {
	prompter := &scriptedRemoveConfirmationPrompter{cancelled: true}

	if outcome := ConfirmRemove("work", prompter); outcome != RemoveConfirmationCancelled {
		t.Fatalf("ConfirmRemove() = %v, want cancelled", outcome)
	}
}

type scriptedRemoveConfirmationPrompter struct {
	confirmed bool
	cancelled bool
	question  ConfirmPrompt
}

func (prompter *scriptedRemoveConfirmationPrompter) Confirm(question ConfirmPrompt) (bool, bool) {
	prompter.question = question
	return prompter.confirmed, prompter.cancelled
}
