package pulse

import (
	"context"
	"errors"
	"time"
)

var errPulseGitUnavailable = errors.New("Git is not installed or not available in the system PATH.")

// Presenter owns git pulse's semantic terminal events.
type Presenter interface {
	Introduction(string)
	RepositoriesFound(int)
	NoRepositories()
	NoCommits()
	Cancelled()
	Present(Report)
}

// Dependencies are the command-owned boundaries required by git pulse.
type Dependencies struct {
	WorkingDirectory func() (string, error)
	Stater           PathStater
	Reader           DirectoryReader
	Yield            func()
	Git              GitRunner
	Prompter         Prompter
	Presenter        Presenter
	Tracker          Tracker
	Now              func() time.Time
}

// Result records one completed git pulse command outcome.
type Result struct {
	Report             Report
	FailedRepositories int
}

// Module owns git pulse behavior behind its typed command interface.
type Module struct {
	workingDirectory func() (string, error)
	stater           PathStater
	reader           DirectoryReader
	yield            func()
	git              GitRunner
	prompter         Prompter
	presenter        Presenter
	tracker          Tracker
	now              func() time.Time
}

// New constructs a git pulse command module.
func New(dependencies Dependencies) (*Module, error) {
	if dependencies.WorkingDirectory == nil {
		return nil, errors.New("git pulse working directory is required")
	}
	if dependencies.Stater == nil {
		return nil, errors.New("git pulse path stater is required")
	}
	if dependencies.Reader == nil {
		return nil, errors.New("git pulse directory reader is required")
	}
	if dependencies.Yield == nil {
		return nil, errors.New("git pulse scheduler yield is required")
	}
	if dependencies.Git == nil {
		return nil, errors.New("git pulse runner is required")
	}
	if dependencies.Prompter == nil {
		return nil, errors.New("git pulse prompter is required")
	}
	if dependencies.Presenter == nil {
		return nil, errors.New("git pulse presenter is required")
	}
	if dependencies.Tracker == nil {
		return nil, errors.New("git pulse tracker is required")
	}
	if dependencies.Now == nil {
		return nil, errors.New("git pulse clock is required")
	}
	return &Module{
		workingDirectory: dependencies.WorkingDirectory,
		stater:           dependencies.Stater,
		reader:           dependencies.Reader,
		yield:            dependencies.Yield,
		git:              dependencies.Git,
		prompter:         dependencies.Prompter,
		presenter:        dependencies.Presenter,
		tracker:          dependencies.Tracker,
		now:              dependencies.Now,
	}, nil
}

// Run executes git pulse without owning process exit behavior.
func (module *Module) Run(ctx context.Context, input Input) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	workingDirectory, err := module.workingDirectory()
	if err != nil {
		return Result{}, err
	}
	root, err := resolvePulseRoot(workingDirectory, input.Directory, module.stater)
	if err != nil {
		return Result{}, err
	}
	if err := verifyPulseGit(ctx, module.git); err != nil {
		return Result{}, err
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	module.presenter.Introduction(root)
	found := 0
	var repositories []string
	err = module.track(ctx, Phase{Kind: PhaseScan, State: PhaseActive, Root: root}, func(report func(Phase)) error {
		var scanErr error
		repositories, scanErr = ScanRepositories(ctx, root, module.reader, func(repository string) {
			found++
			report(Phase{Kind: PhaseScan, State: PhaseActive, Root: root, Repository: repository, Completed: found})
		}, module.yield)
		return scanErr
	})
	if err != nil {
		return Result{}, err
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if len(repositories) == 0 {
		module.presenter.NoRepositories()
		return Result{}, nil
	}
	module.presenter.RepositoriesFound(len(repositories))

	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	days, cancelled, err := selectDays(input, module.prompter)
	if err != nil {
		return Result{}, err
	}
	if cancelled {
		module.presenter.Cancelled()
		return Result{}, nil
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	var fetched FetchResult
	err = module.track(ctx, Phase{Kind: PhaseFetch, State: PhaseActive, Root: root, Total: len(repositories)}, func(report func(Phase)) error {
		var fetchErr error
		fetched, fetchErr = FetchCommits(ctx, module.git, repositories, SinceBoundary(module.now(), days), func(repository string, done int) {
			report(Phase{Kind: PhaseFetch, State: PhaseActive, Root: root, Repository: repository, Completed: done, Total: len(repositories)})
		})
		return fetchErr
	})
	if err != nil {
		return Result{}, err
	}
	if len(fetched.Commits) == 0 {
		module.presenter.NoCommits()
		return Result{FailedRepositories: fetched.FailedRepositories}, nil
	}

	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	commits, cancelled, err := SelectAuthors(fetched.Commits, module.prompter)
	if err != nil {
		return Result{}, err
	}
	if cancelled {
		module.presenter.Cancelled()
		return Result{FailedRepositories: fetched.FailedRepositories}, nil
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	report := BuildReport(commits)
	module.presenter.Present(report)
	return Result{Report: report, FailedRepositories: fetched.FailedRepositories}, nil
}

func verifyPulseGit(ctx context.Context, runner GitRunner) error {
	output, err := runner.Run(ctx, []string{"--version"})
	if err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return err
		}
		return errPulseGitUnavailable
	}
	if output.ExitCode != 0 {
		return errPulseGitUnavailable
	}
	return nil
}
