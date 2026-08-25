package cliapp

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/hackycy/hackycy-cli/internal/logging"
)

func TestUpgradeBindingUsesTypedHandlerAndNoArguments(t *testing.T) {
	called := false
	app, err := New(BuildInfo{Version: "1.0.0"}, Dependencies{
		Out:     &strings.Builder{},
		Err:     &strings.Builder{},
		Logging: logging.NewRuntime(logging.Options{Writer: &strings.Builder{}}),
		Upgrade: func(context.Context) error {
			called = true
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if outcome := app.Execute(context.Background(), []string{"upgrade"}); outcome.Code != 0 || !called {
		t.Fatalf("upgrade outcome = %#v, called=%v", outcome, called)
	}
	if outcome := app.Execute(context.Background(), []string{"upgrade", "extra"}); outcome.Code == 0 {
		t.Fatal("upgrade accepted a positional argument")
	}
}

func TestUpgradeHandlerPreservesTypedExitCode(t *testing.T) {
	app, err := New(BuildInfo{Version: "1.0.0"}, Dependencies{
		Out:     &strings.Builder{},
		Err:     &strings.Builder{},
		Logging: logging.NewRuntime(logging.Options{Writer: &strings.Builder{}}),
		Upgrade: func(context.Context) error {
			return &testExitError{code: 0, err: errors.New("fixture abort")}
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if outcome := app.Execute(context.Background(), []string{"upgrade"}); outcome.Code != 0 {
		t.Fatalf("typed exit outcome = %#v", outcome)
	}
}

type testExitError struct {
	code int
	err  error
}

func (err *testExitError) Error() string { return err.err.Error() }
func (err *testExitError) ExitCode() int { return err.code }
