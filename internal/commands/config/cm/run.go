package cm

import (
	"context"
	"errors"
	"io"
)

// Dependencies are the command-owned external adapters.
type Dependencies struct {
	Reader Reader
	Output io.Writer
}

// Module owns config cm list behavior behind one typed command interface.
type Module struct {
	reader Reader
	output io.Writer
}

// New constructs a config cm list command module.
func New(dependencies Dependencies) (*Module, error) {
	if dependencies.Reader == nil {
		return nil, errors.New("config cm reader is required")
	}
	if dependencies.Output == nil {
		return nil, errors.New("config cm output is required")
	}
	return &Module{reader: dependencies.Reader, output: dependencies.Output}, nil
}

// Run lists configured CM profiles without mutating configuration.
func (module *Module) Run(_ context.Context, _ Input) (Result, error) {
	profiles, err := Read(module.reader)
	if err != nil {
		return Result{}, err
	}
	if _, err := io.WriteString(module.output, Render(profiles)); err != nil {
		return Result{}, err
	}
	return Result{}, nil
}
