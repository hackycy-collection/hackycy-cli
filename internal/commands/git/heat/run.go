package heat

import (
	"context"
	"errors"
	"time"
)

// Dependencies are the command-owned boundaries required by git heat.
type Dependencies struct {
	Git GitRunner
	Now func() time.Time
}

// Result records one completed git heat command outcome.
type Result struct {
	Report Report
	Now    time.Time
}

// Module owns git heat behavior behind its typed command interface.
type Module struct {
	git GitRunner
	now func() time.Time
}

// New constructs a git heat command module.
func New(dependencies Dependencies) (*Module, error) {
	if dependencies.Git == nil {
		return nil, errors.New("git heat runner is required")
	}
	if dependencies.Now == nil {
		return nil, errors.New("git heat clock is required")
	}
	return &Module{
		git: dependencies.Git,
		now: dependencies.Now,
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
	return Result{Report: report, Now: now}, nil
}
