package fork

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
)

// Input is the typed Git Fork command request.
type Input struct {
	Repository  string
	Destination string
}

// DirectoryReader checks whether the requested destination already has entries.
type DirectoryReader interface {
	ReadDir(string) ([]fs.DirEntry, error)
}

// OverwritePrompt describes the only mutation confirmation in Git Fork.
type OverwritePrompt struct {
	Destination string
	Message     string
}

// OverwritePrompter obtains the legacy default-yes replacement decision.
type OverwritePrompter interface {
	ConfirmOverwrite(OverwritePrompt) (confirmed bool, cancelled bool)
}

// Provider resolves a default branch and downloads a provider archive.
type Provider interface {
	DefaultBranch(context.Context, Repository) (string, error)
	DownloadArchive(context.Context, Repository, string) ([]byte, error)
}

// Presenter owns the observable terminal milestones of a Git Fork operation.
type Presenter interface {
	Introduction()
	Resolved(Repository)
	DefaultBranchStarted()
	DefaultBranchResolved(string)
	DefaultBranchFailed(error)
	ArchiveStarted()
	ArchiveSucceeded()
	ArchiveFailed(error)
	CloneStarted()
	CloneSucceeded()
	Cancelled()
	Completed(string)
}

// Acquisition identifies the successful source of the working tree.
type Acquisition string

const (
	acquisitionArchive Acquisition = "archive"
	acquisitionClone   Acquisition = "clone"
)

// Result records Git Fork's normal completion or prompt-cancellation outcome.
type Result struct {
	Cancelled          bool
	Repository         Repository
	Destination        string
	DestinationPath    string
	Ref                string
	Acquisition        Acquisition
	DefaultBranchError error
	ArchiveError       error
}

// Dependencies are the command-owned collaboration points for Git Fork orchestration.
type Dependencies struct {
	Config           ConfigReader
	WorkingDirectory func() (string, error)
	Directories      DirectoryReader
	Prompter         OverwritePrompter
	Provider         Provider
	Extractor        ArchiveExtractor
	CloneRunner      CloneRunner
	Remover          DirectoryRemover
	Presenter        Presenter
}

// Module owns Git Fork's destination mutation and archive-or-clone behavior.
type Module struct {
	config           ConfigReader
	workingDirectory func() (string, error)
	directories      DirectoryReader
	prompter         OverwritePrompter
	provider         Provider
	extractor        ArchiveExtractor
	cloneRunner      CloneRunner
	remover          DirectoryRemover
	presenter        Presenter
}

// New constructs an unregistered Git Fork module.
func New(dependencies Dependencies) (*Module, error) {
	if dependencies.Config == nil {
		return nil, errors.New("git fork config reader is required")
	}
	if dependencies.WorkingDirectory == nil {
		return nil, errors.New("git fork working directory is required")
	}
	if dependencies.Directories == nil {
		return nil, errors.New("git fork directory reader is required")
	}
	if dependencies.Prompter == nil {
		return nil, errors.New("git fork overwrite prompter is required")
	}
	if dependencies.Provider == nil {
		return nil, errors.New("git fork provider is required")
	}
	if dependencies.Extractor == nil {
		return nil, errors.New("git fork archive extractor is required")
	}
	if dependencies.CloneRunner == nil {
		return nil, errors.New("git fork clone runner is required")
	}
	if dependencies.Remover == nil {
		return nil, errors.New("git fork directory remover is required")
	}
	if dependencies.Presenter == nil {
		return nil, errors.New("git fork presenter is required")
	}
	return &Module{
		config:           dependencies.Config,
		workingDirectory: dependencies.WorkingDirectory,
		directories:      dependencies.Directories,
		prompter:         dependencies.Prompter,
		provider:         dependencies.Provider,
		extractor:        dependencies.Extractor,
		cloneRunner:      dependencies.CloneRunner,
		remover:          dependencies.Remover,
		presenter:        dependencies.Presenter,
	}, nil
}

// Run executes Git Fork without owning terminal presentation or process exit status.
func (module *Module) Run(ctx context.Context, input Input) (Result, error) {
	module.presenter.Introduction()
	repository, err := ResolveRepository(input.Repository, module.config)
	if err != nil {
		return Result{}, err
	}
	module.presenter.Resolved(repository)
	workingDirectory, err := module.workingDirectory()
	if err != nil {
		return Result{}, err
	}
	destination := input.Destination
	if destination == "" {
		destination = repository.Name
	}
	destinationPath := resolveDestination(workingDirectory, destination)
	result := Result{Repository: repository, Destination: destination, DestinationPath: destinationPath, Ref: repository.Ref}

	if entries, err := module.directories.ReadDir(destinationPath); err == nil && len(entries) > 0 {
		confirmed, cancelled := module.prompter.ConfirmOverwrite(OverwritePrompt{
			Destination: destination,
			Message:     "Directory \"" + destination + "\" is not empty. Overwrite?",
		})
		if cancelled || !confirmed {
			result.Cancelled = true
			module.presenter.Cancelled()
			return result, nil
		}
		if err := module.remover.RemoveAll(destinationPath); err != nil {
			return Result{}, err
		}
	}

	if result.Ref == "" {
		module.presenter.DefaultBranchStarted()
		result.Ref, result.DefaultBranchError = module.provider.DefaultBranch(ctx, repository)
		if result.DefaultBranchError != nil {
			module.presenter.DefaultBranchFailed(result.DefaultBranchError)
		} else {
			module.presenter.DefaultBranchResolved(result.Ref)
		}
	}
	if result.Ref != "" {
		module.presenter.ArchiveStarted()
		archive, archiveErr := module.provider.DownloadArchive(ctx, repository, result.Ref)
		if archiveErr == nil {
			archiveErr = module.extractor.Extract(destinationPath, archive)
		}
		if archiveErr == nil {
			result.Acquisition = acquisitionArchive
			module.presenter.ArchiveSucceeded()
			module.presenter.Completed(destination)
			return result, nil
		}
		result.ArchiveError = archiveErr
		module.presenter.ArchiveFailed(archiveErr)
	}
	module.presenter.CloneStarted()
	if err := CloneFallback(ctx, module.cloneRunner, module.remover, repositoryWithRef(repository, result.Ref), destinationPath); err != nil {
		return Result{}, err
	}
	result.Acquisition = acquisitionClone
	module.presenter.CloneSucceeded()
	module.presenter.Completed(destination)
	return result, nil
}

func resolveDestination(workingDirectory, destination string) string {
	if filepath.IsAbs(destination) {
		return filepath.Clean(destination)
	}
	return filepath.Clean(filepath.Join(workingDirectory, destination))
}

func repositoryWithRef(repository Repository, ref string) Repository {
	repository.Ref = ref
	return repository
}
