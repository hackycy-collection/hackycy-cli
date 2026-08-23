package cm

import "testing"

func TestConfirmRemoveUsesLegacyQuestionAndAcceptsConfirmation(t *testing.T) {
	prompter := &scriptedCMRemoveConfirmationPrompter{confirmed: true}

	outcome := ConfirmRemove("work", prompter)

	if outcome != RemoveConfirmed {
		t.Fatalf("ConfirmRemove() = %v, want confirmed", outcome)
	}
	if got, want := prompter.question, (RemoveConfirmPrompt{Message: "Remove CM profile \"work\"?"}); got != want {
		t.Fatalf("confirmation question = %#v, want %#v", got, want)
	}
}

func TestConfirmRemoveDoesNotEscapeAConfiguredProfileName(t *testing.T) {
	prompter := &scriptedCMRemoveConfirmationPrompter{confirmed: true}

	ConfirmRemove("work\"profile", prompter)

	if got, want := prompter.question, (RemoveConfirmPrompt{Message: "Remove CM profile \"work\"profile\"?"}); got != want {
		t.Fatalf("confirmation question = %#v, want %#v", got, want)
	}
}

func TestConfirmRemoveRetainsNegativeConfirmation(t *testing.T) {
	prompter := &scriptedCMRemoveConfirmationPrompter{}

	if outcome := ConfirmRemove("work", prompter); outcome != RemoveDeclined {
		t.Fatalf("ConfirmRemove() = %v, want declined", outcome)
	}
}

func TestConfirmRemoveRetainsCancellation(t *testing.T) {
	prompter := &scriptedCMRemoveConfirmationPrompter{cancelled: true}

	if outcome := ConfirmRemove("work", prompter); outcome != RemoveConfirmationCancelled {
		t.Fatalf("ConfirmRemove() = %v, want cancelled", outcome)
	}
}

type scriptedCMRemoveConfirmationPrompter struct {
	confirmed bool
	cancelled bool
	question  RemoveConfirmPrompt
}

func (prompter *scriptedCMRemoveConfirmationPrompter) Confirm(question RemoveConfirmPrompt) (bool, bool) {
	prompter.question = question
	return prompter.confirmed, prompter.cancelled
}
