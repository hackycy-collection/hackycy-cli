package cm

import "fmt"

// RemoveConfirmPrompt describes the confirmation preceding a CM profile removal.
type RemoveConfirmPrompt struct {
	Message string
}

// RemoveConfirmationPrompter confirms one requested CM profile removal.
type RemoveConfirmationPrompter interface {
	Confirm(RemoveConfirmPrompt) (confirmed bool, cancelled bool)
}

// RemoveConfirmation records the read-only decision before a CM profile deletion.
type RemoveConfirmation uint8

const (
	RemoveDeclined RemoveConfirmation = iota
	RemoveConfirmed
	RemoveConfirmationCancelled
)

// ConfirmRemove asks for the legacy CM profile removal confirmation without mutating configuration.
func ConfirmRemove(name string, prompter RemoveConfirmationPrompter) RemoveConfirmation {
	confirmed, cancelled := prompter.Confirm(RemoveConfirmPrompt{Message: fmt.Sprintf("Remove CM profile \"%s\"?", name)})
	if cancelled {
		return RemoveConfirmationCancelled
	}
	if !confirmed {
		return RemoveDeclined
	}
	return RemoveConfirmed
}
