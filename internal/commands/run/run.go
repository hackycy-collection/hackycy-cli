package run

import (
	"context"
	"errors"
	"path/filepath"
)

// Dependencies are the command-owned boundaries required by run.
type Dependencies struct {
	WorkingDirectory func() (string, error)
	Reader           FileReader
	Exists           FileExists
	Prompter         Prompter
	Runner           ChildRunner
	Presenter        Presenter
}

// Module owns run behavior behind its typed command interface.
type Module struct {
	workingDirectory func() (string, error)
	reader           FileReader
	exists           FileExists
	prompter         Prompter
	runner           ChildRunner
	presenter        Presenter
}

// New constructs a run command module.
func New(dependencies Dependencies) (*Module, error) {
	if dependencies.WorkingDirectory == nil {
		return nil, errors.New("run working directory is required")
	}
	if dependencies.Reader == nil {
		return nil, errors.New("run package reader is required")
	}
	if dependencies.Exists == nil {
		return nil, errors.New("run path existence checker is required")
	}
	if dependencies.Prompter == nil {
		return nil, errors.New("run prompter is required")
	}
	if dependencies.Runner == nil {
		return nil, errors.New("run child runner is required")
	}
	if dependencies.Presenter == nil {
		return nil, errors.New("run presenter is required")
	}
	return &Module{
		workingDirectory: dependencies.WorkingDirectory,
		reader:           dependencies.Reader,
		exists:           dependencies.Exists,
		prompter:         dependencies.Prompter,
		runner:           dependencies.Runner,
		presenter:        dependencies.Presenter,
	}, nil
}

// Run executes one package script selection without owning process exit behavior.
func (module *Module) Run(context context.Context, input Input) (Result, error) {
	if err := context.Err(); err != nil {
		return Result{}, err
	}
	workingDirectory, err := module.workingDirectory()
	if err != nil {
		return Result{}, err
	}
	workingDirectory, err = filepath.Abs(workingDirectory)
	if err != nil {
		return Result{}, err
	}
	discovery, err := DiscoverProject(workingDirectory, input.Directory, module.reader)
	if err != nil {
		return Result{}, err
	}

	presentIntroduction(module.presenter)
	script, cancelled, err := selectScript(module.prompter, discovery.Scripts)
	if err != nil {
		return Result{}, err
	}
	if cancelled {
		presentCancellation(module.presenter)
		return Result{}, nil
	}
	if err := context.Err(); err != nil {
		return Result{}, err
	}
	managers, err := PackageManagerOrder(discovery.Directory, module.exists)
	if err != nil {
		return Result{}, err
	}
	manager, cancelled, err := selectPackageManager(module.prompter, managers)
	if err != nil {
		return Result{}, err
	}
	if cancelled {
		presentCancellation(module.presenter)
		return Result{}, nil
	}
	if err := context.Err(); err != nil {
		return Result{}, err
	}
	request := childRequest(discovery.Directory, manager, script)
	presentLaunch(module.presenter, request)
	return module.runner.Run(context, request)
}
