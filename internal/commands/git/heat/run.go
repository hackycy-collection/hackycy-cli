package heat

import (
	"context"
	"errors"
	"time"
)

// Presenter owns the final command report presentation boundary.
type Presenter interface {
	Present(Report, time.Time) error
}

// Dependencies are the command-owned boundaries required by git heat.
type Dependencies struct {
	Git       GitRunner
	Presenter Presenter
	Now       func() time.Time
}

// Result records one completed git heat command outcome.
type Result struct {
	Report Report
}

// Module owns git heat behavior behind its typed command interface.
type Module struct {
	git       GitRunner
	presenter Presenter
	now       func() time.Time
}

// New constructs a git heat command module.
func New(dependencies Dependencies) (*Module, error) {
	if dependencies.Git == nil {
		return nil, errors.New("git heat runner is required")
	}
	if dependencies.Presenter == nil {
		return nil, errors.New("git heat presenter is required")
	}
	if dependencies.Now == nil {
		return nil, errors.New("git heat clock is required")
	}
	return &Module{
		git:       dependencies.Git,
		presenter: dependencies.Presenter,
		now:       dependencies.Now,
	}, nil
}

// Run executes git heat without owning process exit behavior.
func (module *Module) Run(context context.Context, input Input) (Result, error) {
	if err := context.Err(); err != nil {
		return Result{}, err
	}
	options, err := NormalizeInput(input)
	if err != nil {
		return Result{}, err
	}
	repositoryRoot, err := DiscoverRepository(context, module.git)
	if err != nil {
		return Result{}, err
	}
	if err := context.Err(); err != nil {
		return Result{}, err
	}
	log, err := ReadLog(context, module.git, repositoryRoot, options.Range)
	if err != nil {
		return Result{}, err
	}
	if err := context.Err(); err != nil {
		return Result{}, err
	}
	report := BuildReport(repositoryRoot, options, log)
	now := module.now()
	if err := module.presenter.Present(report, now); err != nil {
		return Result{}, err
	}
	return Result{Report: report}, nil
}
