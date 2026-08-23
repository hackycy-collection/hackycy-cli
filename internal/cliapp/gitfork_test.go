package cliapp

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	forkcommand "github.com/hackycy/hackycy-cli/internal/commands/git/fork"
	"github.com/hackycy/hackycy-cli/internal/logging"
)

func TestGitForkBindingPassesItsTwoPositionalArgumentsAndGlobalLogLevel(t *testing.T) {
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	runtime := logging.NewRuntime(logging.Options{Writer: errors})
	var inputs []forkcommand.Input
	app, err := New(BuildInfo{Version: "0.0.0-dev"}, Dependencies{
		Out:     output,
		Err:     errors,
		Logging: runtime,
		GitFork: func(_ context.Context, input forkcommand.Input) (forkcommand.Result, error) {
			inputs = append(inputs, input)
			return forkcommand.Result{}, nil
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	for _, arguments := range [][]string{
		{"git", "fork", "owner/project"},
		{"--log-level", "warn", "git", "fork", "fixture:group/project", "chosen"},
	} {
		outcome := app.Execute(context.Background(), arguments)
		if outcome.Code != 0 || outcome.Err != nil {
			t.Fatalf("%v outcome = %#v, stderr = %q", arguments, outcome, errors.String())
		}
		output.Reset()
		errors.Reset()
	}
	if got, want := inputs, []forkcommand.Input{
		{Repository: "owner/project"},
		{Repository: "fixture:group/project", Destination: "chosen"},
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("inputs = %#v, want %#v", got, want)
	}
	if runtime.Level() != logging.Warn {
		t.Fatalf("log level = %v, want %v", runtime.Level(), logging.Warn)
	}
}

func TestGitForkBindingRejectsInvalidArityBeforeCallingTheHandler(t *testing.T) {
	calls := 0
	app, output, errors, _ := newGitForkTestApp(t, func(context.Context, forkcommand.Input) (forkcommand.Result, error) {
		calls++
		return forkcommand.Result{}, nil
	})
	for _, arguments := range [][]string{
		{"git", "fork"},
		{"git", "fork", "one", "two", "three"},
		{"git", "fork", "owner/project", "--unknown"},
	} {
		outcome := app.Execute(context.Background(), arguments)
		if outcome.Code != 1 || errors.Len() == 0 {
			t.Fatalf("%v outcome = %#v, stderr = %q", arguments, outcome, errors.String())
		}
		output.Reset()
		errors.Reset()
	}
	if calls != 0 {
		t.Fatalf("handler calls = %d", calls)
	}
}

func TestGitForkGroupExposesOnlyForkWhenOtherGitLeavesAreAbsent(t *testing.T) {
	app, output, errors, _ := newGitForkTestApp(t, func(context.Context, forkcommand.Input) (forkcommand.Result, error) {
		return forkcommand.Result{}, nil
	})
	if outcome := app.Execute(context.Background(), []string{"git", "--help"}); outcome.Code != 0 || !strings.Contains(output.String(), "fork") || strings.Contains(output.String(), "heat") || strings.Contains(output.String(), "pulse") {
		t.Fatalf("git help outcome = %#v, stdout = %q", outcome, output.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"git", "heat"}); outcome.Code != 1 || errors.String() != "error: unknown command 'heat'\n" {
		t.Fatalf("absent sibling outcome = %#v, stderr = %q", outcome, errors.String())
	}
}

func TestGitForkBindingPreservesTypedSignalExitWithoutADiagnostic(t *testing.T) {
	app, output, errors, _ := newGitForkTestApp(t, func(context.Context, forkcommand.Input) (forkcommand.Result, error) {
		return forkcommand.Result{}, gitForkExitError{code: 143}
	})
	outcome := app.Execute(context.Background(), []string{"git", "fork", "owner/project"})
	if outcome.Code != 143 || outcome.Err != nil || output.Len() != 0 || errors.Len() != 0 {
		t.Fatalf("outcome = %#v, stdout = %q, stderr = %q", outcome, output.String(), errors.String())
	}
}

func newGitForkTestApp(t *testing.T, handler GitForkHandler) (*App, *bytes.Buffer, *bytes.Buffer, *logging.Runtime) {
	t.Helper()
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	runtime := logging.NewRuntime(logging.Options{Writer: errors})
	app, err := New(BuildInfo{Version: "0.0.0-dev"}, Dependencies{
		Out:     output,
		Err:     errors,
		Logging: runtime,
		GitFork: handler,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return app, output, errors, runtime
}

type gitForkExitError struct {
	code int
}

func (err gitForkExitError) Error() string {
	return "git fork signal outcome"
}

func (err gitForkExitError) ExitCode() int {
	return err.code
}
