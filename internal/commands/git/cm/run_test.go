package cm

import (
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
)

func TestModuleRunGeneratesFromTheResolvedScopeAndSafeProfileProjection(t *testing.T) {
	root := t.TempDir()
	writeStageFile(t, root, "src/value.go")
	runner := newModuleRunner(root, "?? src/value.go\x00")
	resolver := &recordingCMResolver{profile: commitMessageProfile()}
	transport := &recordingProviderTransport{response: &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"choices":[{"message":{"content":"feat(cm): generate a message"}}],"usage":{"prompt_tokens":3}}`)),
	}}
	module := newModule(t, runner, resolver, transport)
	timeout := 1234.0

	result, err := module.Run(context.Background(), Input{Profile: "work", TimeoutMS: &timeout})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.RepositoryRoot != root || result.Scope != ScopeAllUncommitted || result.Cancelled || result.NoChanges || result.Generated == nil || result.Generated.Message != "feat(cm): generate a message" {
		t.Fatalf("result = %#v", result)
	}
	if result.Profile != (ProfileDiagnostic{Name: "work", BaseURL: "https://provider.test/v1", Model: "provider-model"}) {
		t.Fatalf("profile = %#v", result.Profile)
	}
	if resolver.options != (appconfig.CMResolveOptions{ProfileName: "work", TimeoutOverrideMS: &timeout}) {
		t.Fatalf("resolver options = %#v", resolver.options)
	}
	if transport.calls != 1 {
		t.Fatalf("provider calls = %d", transport.calls)
	}
}

func TestModuleRunAppliesStagingModesBeforeCapturingTheSnapshot(t *testing.T) {
	root := t.TempDir()
	writeStageFile(t, root, "value.go")
	t.Run("selection", func(t *testing.T) {
		runner := newModuleRunner(root, "M  value.go\x00")
		module := newModuleWithPrompter(t, runner, &recordingCMResolver{profile: commitMessageProfile()}, successfulProviderTransport(), &stagePrompter{useInitialValues: true})

		result, err := module.Run(context.Background(), Input{Stage: true})
		if err != nil || result.Scope != ScopeStaged || result.Generated == nil {
			t.Fatalf("Run() = (%#v, %v)", result, err)
		}
		requireSnapshotCalls(t, runner,
			[]string{"rev-parse", "--show-toplevel"},
			[]string{"-C", root, "status", "--porcelain=v1", "-z", "--untracked-files=all"},
			[]string{"-C", root, "ls-files", "--stage", "-z"},
			[]string{"-C", root, "add", "-A", "--", "value.go"},
		)
		runner.requireNoCall(t, []string{"-C", root, "diff", "--no-ext-diff", "--find-renames", "--unified=0"})
	})
	t.Run("all", func(t *testing.T) {
		runner := newModuleRunner(root, "M  value.go\x00")
		module := newModule(t, runner, &recordingCMResolver{profile: commitMessageProfile()}, successfulProviderTransport())

		result, err := module.Run(context.Background(), Input{StageAll: true})
		if err != nil || result.Scope != ScopeStaged || result.Generated == nil {
			t.Fatalf("Run() = (%#v, %v)", result, err)
		}
		requireSnapshotCalls(t, runner,
			[]string{"rev-parse", "--show-toplevel"},
			[]string{"-C", root, "add", "-A"},
		)
		runner.requireNoCall(t, []string{"-C", root, "diff", "--no-ext-diff", "--find-renames", "--unified=0"})
	})
}

func TestModuleRunConfirmsThenRechecksAndCommits(t *testing.T) {
	root := t.TempDir()
	writeStageFile(t, root, "value.go")
	t.Run("declined confirmation keeps the generated message and index state", func(t *testing.T) {
		runner := newModuleRunner(root, "M  value.go\x00")
		committer := &recordingCommitPrompter{}
		module := newModuleWithPrompts(t, runner, &recordingCMResolver{profile: commitMessageProfile()}, successfulProviderTransport(), &stagePrompter{}, committer)

		result, err := module.Run(context.Background(), Input{Staged: true})
		if err != nil || !result.Cancelled || result.Committed || result.Generated == nil {
			t.Fatalf("Run() = (%#v, %v)", result, err)
		}
		if committer.prompt.Message != "Create this commit?" || committer.prompt.Generated.Message != result.Generated.Message || committer.prompt.Profile != result.Profile {
			t.Fatalf("prompt = %#v, result = %#v", committer.prompt, result)
		}
		runner.requireNoCall(t, []string{"-C", root, "commit", "-m", result.Generated.Message})
	})
	t.Run("confirmed confirmation commits after a second snapshot", func(t *testing.T) {
		runner := newModuleRunner(root, "M  value.go\x00")
		committer := &recordingCommitPrompter{confirmed: true}
		module := newModuleWithPrompts(t, runner, &recordingCMResolver{profile: commitMessageProfile()}, successfulProviderTransport(), &stagePrompter{}, committer)

		result, err := module.Run(context.Background(), Input{Staged: true})
		if err != nil || result.Cancelled || !result.Committed || result.Generated == nil {
			t.Fatalf("Run() = (%#v, %v)", result, err)
		}
		requireSnapshotCalls(t, runner, []string{"-C", root, "commit", "-m", result.Generated.Message})
	})
}

func TestModuleRunReportsTheCommittedPartialResultWhenPushFails(t *testing.T) {
	root := t.TempDir()
	writeStageFile(t, root, "value.go")
	runner := newModuleRunner(root, "M  value.go\x00")
	runner.responses[snapshotRunnerKey([]string{"-C", root, "branch", "--show-current"})] = GitOutput{Stdout: []byte("main\n")}
	runner.responses[snapshotRunnerKey([]string{"-C", root, "push", "-u", "origin", "main"})] = GitOutput{Stderr: []byte("remote rejected\n"), ExitCode: 1}
	module := newModuleWithPrompts(t, runner, &recordingCMResolver{profile: commitMessageProfile()}, successfulProviderTransport(), &stagePrompter{}, &recordingCommitPrompter{confirmed: true})

	result, err := module.Run(context.Background(), Input{Staged: true, Push: modeString("origin")})
	if err == nil || err.Error() != "remote rejected" || !result.Committed || result.Pushed || result.PushRemote != "" || result.Generated == nil {
		t.Fatalf("Run() = (%#v, %v)", result, err)
	}
}

func TestModuleRunMarksASuccessfulPush(t *testing.T) {
	root := t.TempDir()
	writeStageFile(t, root, "value.go")
	runner := newModuleRunner(root, "M  value.go\x00")
	runner.responses[snapshotRunnerKey([]string{"-C", root, "branch", "--show-current"})] = GitOutput{Stdout: []byte("main\n")}
	module := newModuleWithPrompts(t, runner, &recordingCMResolver{profile: commitMessageProfile()}, successfulProviderTransport(), &stagePrompter{}, &recordingCommitPrompter{confirmed: true})

	result, err := module.Run(context.Background(), Input{Staged: true, Push: modeString("origin")})
	if err != nil || !result.Committed || !result.Pushed || result.PushRemote != "origin" {
		t.Fatalf("Run() = (%#v, %v)", result, err)
	}
}

func TestModuleRunReturnsNormalNoChangeAndSelectionCancellationOutcomesBeforeProfileResolution(t *testing.T) {
	t.Run("no changes", func(t *testing.T) {
		root := t.TempDir()
		resolver := &recordingCMResolver{profile: commitMessageProfile()}
		module := newModule(t, newModuleRunner(root, ""), resolver, successfulProviderTransport())

		result, err := module.Run(context.Background(), Input{})
		if err != nil || result != (Result{RepositoryRoot: root, Scope: ScopeAllUncommitted, NoChanges: true, NoChangeScope: ScopeAllUncommitted}) || resolver.calls != 0 {
			t.Fatalf("Run() = (%#v, %v), resolver calls = %d", result, err, resolver.calls)
		}
	})
	t.Run("selection cancellation", func(t *testing.T) {
		root := t.TempDir()
		resolver := &recordingCMResolver{profile: commitMessageProfile()}
		module := newModuleWithPrompter(t, newModuleRunner(root, " M value.go\x00"), resolver, successfulProviderTransport(), &stagePrompter{cancelled: true})

		result, err := module.Run(context.Background(), Input{Stage: true})
		if err != nil || result != (Result{RepositoryRoot: root, Scope: ScopeStaged, Cancelled: true}) || resolver.calls != 0 {
			t.Fatalf("Run() = (%#v, %v), resolver calls = %d", result, err, resolver.calls)
		}
	})
}

func TestModuleRunResolvesProfileBeforeRejectingLanguageAndDoesNotCallProvider(t *testing.T) {
	root := t.TempDir()
	writeStageFile(t, root, "value.go")
	resolver := &recordingCMResolver{profile: commitMessageProfile()}
	transport := &recordingProviderTransport{}
	module := newModule(t, newModuleRunner(root, "?? value.go\x00"), resolver, transport)

	result, err := module.Run(context.Background(), Input{Language: "fr"})
	if err == nil || err.Error() != "Unsupported language. Use \"en\" or \"zh\"." || result != (Result{}) || resolver.calls != 1 || transport.calls != 0 {
		t.Fatalf("Run() = (%#v, %v), resolver=%d, provider=%d", result, err, resolver.calls, transport.calls)
	}
}

func TestModuleRunPropagatesResolverAndProviderFailuresWithoutLeakingTheAPIKey(t *testing.T) {
	root := t.TempDir()
	writeStageFile(t, root, "value.go")
	t.Run("resolver", func(t *testing.T) {
		failure := errors.New("no usable profile")
		module := newModule(t, newModuleRunner(root, "?? value.go\x00"), &recordingCMResolver{err: failure}, successfulProviderTransport())
		_, err := module.Run(context.Background(), Input{})
		if !errors.Is(err, failure) {
			t.Fatalf("Run() error = %v", err)
		}
	})
	t.Run("provider", func(t *testing.T) {
		profile := commitMessageProfile()
		module := newModule(t, newModuleRunner(root, "?? value.go\x00"), &recordingCMResolver{profile: profile}, providerTransportFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("provider-api-key rejected")
		}))
		result, err := module.Run(context.Background(), Input{})
		if err == nil || strings.Contains(err.Error(), profile.APIKey) || !strings.Contains(err.Error(), "[REDACTED]") || result.Profile != profileDiagnostic(profile) {
			t.Fatalf("Run() = (%#v, %v)", result, err)
		}
	})
}

func TestNewModuleRequiresEveryCommandOwnedDependency(t *testing.T) {
	runner := newModuleRunner(t.TempDir(), "")
	resolver := &recordingCMResolver{profile: commitMessageProfile()}
	transport := successfulProviderTransport()
	for _, dependencies := range []Dependencies{
		{Files: diskSnapshotFileSystem{}, Prompter: &stagePrompter{}, Committer: &recordingCommitPrompter{}, Resolver: resolver, Transport: transport},
		{Git: runner, Prompter: &stagePrompter{}, Committer: &recordingCommitPrompter{}, Resolver: resolver, Transport: transport},
		{Git: runner, Files: diskSnapshotFileSystem{}, Committer: &recordingCommitPrompter{}, Resolver: resolver, Transport: transport},
		{Git: runner, Files: diskSnapshotFileSystem{}, Prompter: &stagePrompter{}, Resolver: resolver, Transport: transport},
		{Git: runner, Files: diskSnapshotFileSystem{}, Prompter: &stagePrompter{}, Committer: &recordingCommitPrompter{}, Transport: transport},
		{Git: runner, Files: diskSnapshotFileSystem{}, Prompter: &stagePrompter{}, Committer: &recordingCommitPrompter{}, Resolver: resolver},
	} {
		if _, err := New(dependencies); err == nil {
			t.Fatalf("New(%#v) error = nil", dependencies)
		}
	}
}

func newModule(t *testing.T, runner *snapshotRunner, resolver ProfileResolver, transport ProviderTransport) *Module {
	t.Helper()
	return newModuleWithPrompts(t, runner, resolver, transport, &stagePrompter{}, &recordingCommitPrompter{})
}

func newModuleWithPrompter(t *testing.T, runner *snapshotRunner, resolver ProfileResolver, transport ProviderTransport, prompter StagePrompter) *Module {
	return newModuleWithPrompts(t, runner, resolver, transport, prompter, &recordingCommitPrompter{})
}

func newModuleWithPrompts(t *testing.T, runner *snapshotRunner, resolver ProfileResolver, transport ProviderTransport, prompter StagePrompter, committer CommitPrompter) *Module {
	t.Helper()
	module, err := New(Dependencies{Git: runner, Files: diskSnapshotFileSystem{}, Prompter: prompter, Committer: committer, Resolver: resolver, Transport: transport})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return module
}

type recordingCommitPrompter struct {
	prompt    CommitPrompt
	confirmed bool
	cancelled bool
}

func (prompter *recordingCommitPrompter) ConfirmCommit(prompt CommitPrompt) (bool, bool) {
	prompter.prompt = prompt
	return prompter.confirmed, prompter.cancelled
}

func newModuleRunner(root, status string) *snapshotRunner {
	runner := newSnapshotRunner(root, status)
	setEmptySnapshotDiffs(runner, root)
	return runner
}

func successfulProviderTransport() ProviderTransport {
	return providerTransportFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"choices":[{"message":{"content":"feat(cm): generate a message"}}]}`))}, nil
	})
}

func requireSnapshotCalls(t *testing.T, runner *snapshotRunner, expected ...[]string) {
	t.Helper()
	runner.mu.Lock()
	defer runner.mu.Unlock()
	for _, want := range expected {
		found := false
		for _, actual := range runner.calls {
			if reflect.DeepEqual(actual, want) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Git calls = %#v, missing %#v", runner.calls, want)
		}
	}
}

type recordingCMResolver struct {
	profile appconfig.ResolvedCMProfile
	options appconfig.CMResolveOptions
	err     error
	calls   int
}

func (resolver *recordingCMResolver) ResolveCMProfile(options appconfig.CMResolveOptions) (appconfig.ResolvedCMProfile, error) {
	resolver.calls++
	resolver.options = options
	return resolver.profile, resolver.err
}
