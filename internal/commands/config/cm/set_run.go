package cm

import (
	"context"
	"errors"
)

// SetDependencies are the command-owned adapters for config cm set.
type SetDependencies struct {
	Writer    SetWriter
	Presenter SetPresenter
}

// SetModule owns config cm set behavior behind its typed command interface.
type SetModule struct {
	writer    SetWriter
	presenter SetPresenter
}

// NewSet constructs a config cm set command module.
func NewSet(dependencies SetDependencies) (*SetModule, error) {
	if dependencies.Writer == nil {
		return nil, errors.New("config cm set writer is required")
	}
	if dependencies.Presenter == nil {
		return nil, errors.New("config cm set presenter is required")
	}
	return &SetModule{writer: dependencies.Writer, presenter: dependencies.Presenter}, nil
}

// Run updates one CM profile field and presents only successful outcomes.
func (module *SetModule) Run(_ context.Context, request SetRequest) (SetResult, error) {
	if err := SetProfileValue(module.writer, request); err != nil {
		return SetResult{}, err
	}
	PresentSetSuccess(module.presenter, request)
	return SetResult{}, nil
}
