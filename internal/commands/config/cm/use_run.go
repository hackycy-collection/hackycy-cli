package cm

import (
	"context"
	"errors"
)

// UseDependencies are the command-owned adapters for config cm use.
type UseDependencies struct {
	Writer    UseWriter
	Presenter UsePresenter
}

// UseModule owns config cm use behavior behind its typed command interface.
type UseModule struct {
	writer    UseWriter
	presenter UsePresenter
}

// NewUse constructs a config cm use command module.
func NewUse(dependencies UseDependencies) (*UseModule, error) {
	if dependencies.Writer == nil {
		return nil, errors.New("config cm use writer is required")
	}
	if dependencies.Presenter == nil {
		return nil, errors.New("config cm use presenter is required")
	}
	return &UseModule{writer: dependencies.Writer, presenter: dependencies.Presenter}, nil
}

// Run selects the requested CM profile and presents the successful outcome.
func (module *UseModule) Run(_ context.Context, request UseRequest) (UseResult, error) {
	if err := SelectDefaultProfile(module.writer, request.Profile); err != nil {
		return UseResult{}, err
	}
	PresentUseSuccess(module.presenter, request.Profile)
	return UseResult{}, nil
}
