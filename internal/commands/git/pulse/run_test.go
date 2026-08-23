package pulse

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestModuleRunComposesScanPromptFetchAuthorSelectionAndReport(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "nested")
	makePulseDirectory(t, filepath.Join(root, ".git"))
	makePulseDirectory(t, filepath.Join(nested, ".git"))
	runner := &modulePulseGitRunner{logs: map[string]pulseGitResponse{
		root:   {output: GitOutput{Stdout: []byte("Ada\x1f2026-08-23 10:00:00\x1froot commit")}},
		nested: {output: GitOutput{Stdout: []byte("Ben\x1f2026-08-22 10:00:00\x1fnested commit")}},
	}}
	prompter := &scriptedPulsePrompter{days: 2, authors: []string{"Ada"}}
	presenter := &recordingPulsePresenter{}
	module := newPulseModuleForTest(t, root, runner, prompter, presenter)

	result, err := module.Run(context.Background(), Input{})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.FailedRepositories != 0 || result.Report.CommitCount != 1 {
		t.Fatalf("result = %#v", result)
	}
	wantReport := BuildReport([]Commit{{Repository: root, Author: "Ada", Date: "2026-08-23 10:00:00", Subject: "root commit"}})
	if !reflect.DeepEqual(result.Report, wantReport) {
		t.Fatalf("report = %#v, want %#v", result.Report, wantReport)
	}
	if got, want := presenter.introductions, []string{root}; !reflect.DeepEqual(got, want) {
		t.Fatalf("introductions = %#v, want %#v", got, want)
	}
	if presenter.scanStarts != 1 || presenter.foundRepositories != 2 || presenter.repositoriesFound != 2 {
		t.Fatalf("scan presentation = %#v", presenter)
	}
	if presenter.fetchStarts != 1 || presenter.fetchProgress != 2 {
		t.Fatalf("fetch presentation = %#v", presenter)
	}
	if !reflect.DeepEqual(presenter.reports, []Report{wantReport}) {
		t.Fatalf("presented reports = %#v", presenter.reports)
	}
	if got := runner.sinceArguments(); !reflect.DeepEqual(got, []string{"--since=2026-08-22 00:00:00", "--since=2026-08-22 00:00:00"}) {
		t.Fatalf("since arguments = %#v", got)
	}
}

func TestModuleRunReturnsSuccessfulNoResultAndPromptCancellationOutcomes(t *testing.T) {
	t.Run("no repositories", func(t *testing.T) {
		root := t.TempDir()
		runner := &modulePulseGitRunner{}
		prompter := &scriptedPulsePrompter{}
		presenter := &recordingPulsePresenter{}
		module := newPulseModuleForTest(t, root, runner, prompter, presenter)

		result, err := module.Run(context.Background(), Input{})
		if err != nil || !reflect.DeepEqual(result, Result{}) {
			t.Fatalf("Run() = (%#v, %v)", result, err)
		}
		if presenter.noRepositories != 1 || runner.logCallCount() != 0 || prompter.dayCalls != 0 {
			t.Fatalf("no-repository outcome = presenter %#v, Git logs %d, day calls %d", presenter, runner.logCallCount(), prompter.dayCalls)
		}
	})

	t.Run("day cancellation", func(t *testing.T) {
		root := t.TempDir()
		makePulseDirectory(t, filepath.Join(root, ".git"))
		runner := &modulePulseGitRunner{}
		prompter := &scriptedPulsePrompter{daysCancelled: true}
		presenter := &recordingPulsePresenter{}
		module := newPulseModuleForTest(t, root, runner, prompter, presenter)

		result, err := module.Run(context.Background(), Input{})
		if err != nil || !reflect.DeepEqual(result, Result{}) || presenter.cancelled != 1 || runner.logCallCount() != 0 {
			t.Fatalf("day cancellation = (%#v, %v), presenter %#v, logs %d", result, err, presenter, runner.logCallCount())
		}
	})

	t.Run("all failed repositories", func(t *testing.T) {
		root := t.TempDir()
		makePulseDirectory(t, filepath.Join(root, ".git"))
		runner := &modulePulseGitRunner{logs: map[string]pulseGitResponse{root: {err: errors.New("failed")}}}
		prompter := &scriptedPulsePrompter{days: 1}
		presenter := &recordingPulsePresenter{}
		module := newPulseModuleForTest(t, root, runner, prompter, presenter)

		result, err := module.Run(context.Background(), Input{})
		if err != nil || result.FailedRepositories != 1 || presenter.noCommits != 1 {
			t.Fatalf("partial no-result = (%#v, %v), presenter %#v", result, err, presenter)
		}
	})

	t.Run("author cancellation", func(t *testing.T) {
		root := t.TempDir()
		makePulseDirectory(t, filepath.Join(root, ".git"))
		runner := &modulePulseGitRunner{logs: map[string]pulseGitResponse{root: {output: GitOutput{Stdout: []byte("Ada\x1f2026-08-23 10:00:00\x1fone\nBen\x1f2026-08-23 09:00:00\x1ftwo")}}}}
		prompter := &scriptedPulsePrompter{days: 1, authorsCancelled: true}
		presenter := &recordingPulsePresenter{}
		module := newPulseModuleForTest(t, root, runner, prompter, presenter)

		result, err := module.Run(context.Background(), Input{})
		if err != nil || !reflect.DeepEqual(result.Report, Report{}) || presenter.cancelled != 1 || len(presenter.reports) != 0 {
			t.Fatalf("author cancellation = (%#v, %v), presenter %#v", result, err, presenter)
		}
	})
}

func TestModuleRunValidatesRootAndGitBeforePresentation(t *testing.T) {
	root := t.TempDir()
	presenter := &recordingPulsePresenter{}
	runner := &modulePulseGitRunner{}
	module := newPulseModuleForTest(t, root, runner, &scriptedPulsePrompter{}, presenter)

	_, err := module.Run(context.Background(), Input{Directory: "missing"})
	if got, want := err.Error(), "Directory not found: "+filepath.Join(root, "missing"); got != want {
		t.Fatalf("missing root error = %q, want %q", got, want)
	}
	if runner.callCount() != 0 || len(presenter.introductions) != 0 {
		t.Fatalf("missing root performed Git or presentation work")
	}

	makePulseDirectory(t, filepath.Join(root, "valid"))
	runner.version = GitOutput{ExitCode: 1}
	_, err = module.Run(context.Background(), Input{Directory: "valid"})
	if !errors.Is(err, errPulseGitUnavailable) || len(presenter.introductions) != 0 {
		t.Fatalf("unavailable Git result = %v, presenter %#v", err, presenter)
	}
}

func TestModuleRunReturnsPreCancelledContextWithoutWork(t *testing.T) {
	root := t.TempDir()
	runner := &modulePulseGitRunner{}
	presenter := &recordingPulsePresenter{}
	module := newPulseModuleForTest(t, root, runner, &scriptedPulsePrompter{}, presenter)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := module.Run(ctx, Input{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want context cancellation", err)
	}
	if runner.callCount() != 0 || len(presenter.introductions) != 0 {
		t.Fatal("pre-cancelled Run() performed work")
	}
}

func TestModuleRunDoesNotPromptAfterContextCancellationAtPromptBoundaries(t *testing.T) {
	t.Run("before day prompt", func(t *testing.T) {
		root := t.TempDir()
		makePulseDirectory(t, filepath.Join(root, ".git"))
		ctx, cancel := context.WithCancel(context.Background())
		reader := directoryReaderFunc(func(path string) ([]os.DirEntry, error) {
			entries, err := os.ReadDir(path)
			if path == root {
				cancel()
			}
			return entries, err
		})
		prompter := &scriptedPulsePrompter{}
		module := newPulseModuleWithDependencies(t, Dependencies{
			WorkingDirectory: func() (string, error) { return root, nil },
			Stater:           osPulseStater{},
			Reader:           reader,
			Yield:            func() {},
			Git:              &modulePulseGitRunner{},
			Prompter:         prompter,
			Presenter:        &recordingPulsePresenter{},
			Now:              time.Now,
		})

		_, err := module.Run(ctx, Input{})
		if !errors.Is(err, context.Canceled) || prompter.dayCalls != 0 {
			t.Fatalf("Run() = (%v, day calls %d)", err, prompter.dayCalls)
		}
	})

	t.Run("before author prompt", func(t *testing.T) {
		root := t.TempDir()
		makePulseDirectory(t, filepath.Join(root, ".git"))
		ctx, cancel := context.WithCancel(context.Background())
		runner := &cancellingAfterPulseLogRunner{root: root, cancel: cancel}
		prompter := &scriptedPulsePrompter{days: 1}
		module := newPulseModuleForTest(t, root, runner, prompter, &recordingPulsePresenter{})

		_, err := module.Run(ctx, Input{})
		if !errors.Is(err, context.Canceled) || prompter.authorPrompt.Message != "" {
			t.Fatalf("Run() = (%v, author prompt %#v)", err, prompter.authorPrompt)
		}
	})
}

func TestModuleRunPreservesTypedSignalErrorFromGitAvailabilityCheck(t *testing.T) {
	root := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	signal := pulseFetchExitError{cause: context.Canceled, code: 143}
	module := newPulseModuleForTest(t, root, pulseGitRunnerFunc(func(_ context.Context, _ []string) (GitOutput, error) {
		cancel()
		return GitOutput{}, signal
	}), &scriptedPulsePrompter{}, &recordingPulsePresenter{})

	_, err := module.Run(ctx, Input{})
	var exitOutcome pulseExitCodedError
	if !errors.As(err, &exitOutcome) || exitOutcome.ExitCode() != 143 {
		t.Fatalf("Run() error = %v, want typed signal outcome", err)
	}
}

func newPulseModuleForTest(t *testing.T, root string, runner GitRunner, prompter Prompter, presenter Presenter) *Module {
	t.Helper()
	return newPulseModuleWithDependencies(t, Dependencies{
		WorkingDirectory: func() (string, error) { return root, nil },
		Stater:           osPulseStater{},
		Reader:           osDirectoryReader{},
		Yield:            func() {},
		Git:              runner,
		Prompter:         prompter,
		Presenter:        presenter,
		Now: func() time.Time {
			return time.Date(2026, time.August, 23, 14, 0, 0, 0, time.UTC)
		},
	})
}

func newPulseModuleWithDependencies(t *testing.T, dependencies Dependencies) *Module {
	t.Helper()
	module, err := New(dependencies)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return module
}

type cancellingAfterPulseLogRunner struct {
	root   string
	cancel context.CancelFunc
}

func (runner *cancellingAfterPulseLogRunner) Run(_ context.Context, arguments []string) (GitOutput, error) {
	if reflect.DeepEqual(arguments, []string{"--version"}) {
		return GitOutput{}, nil
	}
	runner.cancel()
	return GitOutput{Stdout: []byte("Ada\x1f2026-08-23 10:00:00\x1fone\nBen\x1f2026-08-23 09:00:00\x1ftwo")}, nil
}

type pulseGitRunnerFunc func(context.Context, []string) (GitOutput, error)

func (function pulseGitRunnerFunc) Run(context context.Context, arguments []string) (GitOutput, error) {
	return function(context, arguments)
}

type scriptedPulsePrompter struct {
	dayPrompt        DayPrompt
	days             int
	daysCancelled    bool
	dayCalls         int
	authorPrompt     AuthorPrompt
	authors          []string
	authorsCancelled bool
}

func (prompter *scriptedPulsePrompter) SelectDays(prompt DayPrompt) (int, bool) {
	prompter.dayPrompt = prompt
	prompter.dayCalls++
	return prompter.days, prompter.daysCancelled
}

func (prompter *scriptedPulsePrompter) SelectAuthors(prompt AuthorPrompt) ([]string, bool) {
	prompter.authorPrompt = prompt
	return prompter.authors, prompter.authorsCancelled
}

type modulePulseGitRunner struct {
	mu         sync.Mutex
	version    GitOutput
	versionErr error
	logs       map[string]pulseGitResponse
	calls      [][]string
}

func (runner *modulePulseGitRunner) Run(_ context.Context, arguments []string) (GitOutput, error) {
	runner.mu.Lock()
	defer runner.mu.Unlock()
	runner.calls = append(runner.calls, append([]string(nil), arguments...))
	if reflect.DeepEqual(arguments, []string{"--version"}) {
		return runner.version, runner.versionErr
	}
	if len(arguments) < 2 {
		return GitOutput{}, fmt.Errorf("unexpected Git arguments: %q", arguments)
	}
	response := runner.logs[arguments[1]]
	return response.output, response.err
}

func (runner *modulePulseGitRunner) callCount() int {
	runner.mu.Lock()
	defer runner.mu.Unlock()
	return len(runner.calls)
}

func (runner *modulePulseGitRunner) logCallCount() int {
	runner.mu.Lock()
	defer runner.mu.Unlock()
	count := 0
	for _, arguments := range runner.calls {
		if !reflect.DeepEqual(arguments, []string{"--version"}) {
			count++
		}
	}
	return count
}

func (runner *modulePulseGitRunner) sinceArguments() []string {
	runner.mu.Lock()
	defer runner.mu.Unlock()
	arguments := make([]string, 0)
	for _, invocation := range runner.calls {
		for _, argument := range invocation {
			if len(argument) >= len("--since=") && argument[:len("--since=")] == "--since=" {
				arguments = append(arguments, argument)
			}
		}
	}
	return arguments
}

type recordingPulsePresenter struct {
	introductions     []string
	scanStarts        int
	foundRepositories int
	repositoriesFound int
	noRepositories    int
	fetchStarts       int
	fetchProgress     int
	noCommits         int
	cancelled         int
	reports           []Report
}

func (presenter *recordingPulsePresenter) Introduction(root string) {
	presenter.introductions = append(presenter.introductions, root)
}

func (presenter *recordingPulsePresenter) ScanStarted() {
	presenter.scanStarts++
}

func (presenter *recordingPulsePresenter) RepositoryFound(string, string, int) {
	presenter.foundRepositories++
}

func (presenter *recordingPulsePresenter) RepositoriesFound(count int) {
	presenter.repositoriesFound = count
}

func (presenter *recordingPulsePresenter) NoRepositories() {
	presenter.noRepositories++
}

func (presenter *recordingPulsePresenter) FetchStarted(int) {
	presenter.fetchStarts++
}

func (presenter *recordingPulsePresenter) FetchProgress(string, string, int, int) {
	presenter.fetchProgress++
}

func (presenter *recordingPulsePresenter) NoCommits() {
	presenter.noCommits++
}

func (presenter *recordingPulsePresenter) Cancelled() {
	presenter.cancelled++
}

func (presenter *recordingPulsePresenter) Present(report Report) {
	presenter.reports = append(presenter.reports, report)
}
