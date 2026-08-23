package exportenv

import (
	"context"
	"errors"
	"path/filepath"
)

// Input is the typed export env command input.
type Input struct {
	Directory   string
	Environment string
	Merge       bool
	Output      string
}

// Result records a successful command outcome.
type Result struct {
	Cancelled bool
}

// Dependencies are the command-owned external adapters.
type Dependencies struct {
	WorkingDirectory func() (string, error)
	Selector         EnvironmentSelector
	Reader           FileReader
	Writer           FileWriter
	Presenter        Presenter
}

// Module owns export env behavior behind one typed command interface.
type Module struct {
	workingDirectory func() (string, error)
	selector         EnvironmentSelector
	reader           FileReader
	writer           FileWriter
	presenter        Presenter
}

// New constructs an export env command module.
func New(dependencies Dependencies) (*Module, error) {
	if dependencies.WorkingDirectory == nil {
		return nil, errors.New("export env working directory is required")
	}
	if dependencies.Selector == nil {
		return nil, errors.New("export env selector is required")
	}
	if dependencies.Reader == nil {
		return nil, errors.New("export env reader is required")
	}
	if dependencies.Writer == nil {
		return nil, errors.New("export env writer is required")
	}
	if dependencies.Presenter == nil {
		return nil, errors.New("export env presenter is required")
	}
	return &Module{
		workingDirectory: dependencies.WorkingDirectory,
		selector:         dependencies.Selector,
		reader:           dependencies.Reader,
		writer:           dependencies.Writer,
		presenter:        dependencies.Presenter,
	}, nil
}

// Run executes the export env command without owning process exit behavior.
func (module *Module) Run(_ context.Context, input Input) (Result, error) {
	workingDirectory, err := module.workingDirectory()
	if err != nil {
		return Result{}, err
	}
	workingDirectory, err = filepath.Abs(workingDirectory)
	if err != nil {
		return Result{}, err
	}

	discovery, err := Discover(resolveInputDirectory(workingDirectory, input.Directory))
	if err != nil {
		return Result{}, err
	}
	selection, err := Select(discovery, SelectionOptions{
		Environment: input.Environment,
		Merge:       input.Merge,
	}, module.selector)
	if err != nil {
		return Result{}, err
	}
	if selection.Cancelled {
		PresentCancellation(module.presenter)
		return Result{Cancelled: true}, nil
	}

	contents, err := Read(discovery, selection, module.reader)
	if err != nil {
		return Result{}, err
	}
	output, err := Encode(Parse(contents...))
	if err != nil {
		return Result{}, err
	}
	if input.Output != "" {
		Present(module.presenter, output, input.Output)
		if err := WriteOutput(workingDirectory, input.Output, output, module.writer); err != nil {
			return Result{}, err
		}
		return Result{}, nil
	}

	Present(module.presenter, output, "")
	return Result{}, nil
}

func resolveInputDirectory(workingDirectory, directory string) string {
	if directory == "" {
		return workingDirectory
	}
	if filepath.IsAbs(directory) {
		return filepath.Clean(directory)
	}
	return filepath.Join(workingDirectory, directory)
}
