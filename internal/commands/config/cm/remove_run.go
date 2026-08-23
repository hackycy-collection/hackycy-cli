package cm

import (
	"context"
	"errors"
	"fmt"
)

// RemoveRequest is the typed CLI request for removing one CM profile.
type RemoveRequest struct {
	Profile string
}

// RemoveResult records successful non-mutating remove outcomes.
type RemoveResult struct {
	Cancelled bool
	Declined  bool
}

// RemoveDependencies are the command-owned adapters for config cm remove.
type RemoveDependencies struct {
	Prompter  RemoveConfirmationPrompter
	Writer    RemoveWriter
	Presenter RemovePresenter
}

// RemoveModule owns config cm remove behavior behind its typed command interface.
type RemoveModule struct {
	prompter  RemoveConfirmationPrompter
	writer    RemoveWriter
	presenter RemovePresenter
}

// NewRemove constructs a config cm remove command module.
func NewRemove(dependencies RemoveDependencies) (*RemoveModule, error) {
	if dependencies.Prompter == nil {
		return nil, errors.New("config cm remove prompter is required")
	}
	if dependencies.Writer == nil {
		return nil, errors.New("config cm remove writer is required")
	}
	if dependencies.Presenter == nil {
		return nil, errors.New("config cm remove presenter is required")
	}
	return &RemoveModule{
		prompter:  dependencies.Prompter,
		writer:    dependencies.Writer,
		presenter: dependencies.Presenter,
	}, nil
}

// Run confirms, removes, and presents one CM profile removal outcome.
func (module *RemoveModule) Run(_ context.Context, request RemoveRequest) (RemoveResult, error) {
	switch ConfirmRemove(request.Profile, module.prompter) {
	case RemoveConfirmationCancelled:
		PresentRemoveCancellation(module.presenter)
		return RemoveResult{Cancelled: true}, nil
	case RemoveDeclined:
		PresentRemoveCancellation(module.presenter)
		return RemoveResult{Declined: true}, nil
	}

	removed, err := RemoveProfile(module.writer, request.Profile)
	if err != nil {
		return RemoveResult{}, err
	}
	if !removed {
		return RemoveResult{}, fmt.Errorf("CM profile not found: %s", request.Profile)
	}
	PresentRemoveSuccess(module.presenter, request.Profile)
	return RemoveResult{}, nil
}
