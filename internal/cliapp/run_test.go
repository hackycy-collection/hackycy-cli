package cliapp

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	runcommand "github.com/hackycy/hackycy-cli/internal/commands/run"
	"github.com/hackycy/hackycy-cli/internal/logging"
)

func TestRunBinderTreatsOneOptionLikeOperandAsTheLegacyPath(t *testing.T) {
	var inputs []runcommand.Input
	app, output, errors, _ := testApp(t, nil)
	handler := func(_ context.Context, input runcommand.Input) (runcommand.Result, error) {
		inputs = append(inputs, input)
		return runcommand.Result{}, nil
	}

	for _, arguments := range [][]string{{"run"}, {"run", "project"}, {"run", "--flag"}} {
		outcome := executeRunBinder(app, arguments, handler, output, errors)
		if outcome.Code != 0 || outcome.Err != nil {
			t.Fatalf("arguments %q outcome = %#v, stderr = %q", arguments, outcome, errors.String())
		}
		output.Reset()
		errors.Reset()
	}

	want := []runcommand.Input{{}, {Directory: "project"}, {Directory: "--flag"}}
	if !reflect.DeepEqual(inputs, want) {
		t.Fatalf("inputs = %#v, want %#v", inputs, want)
	}
}

func TestRunBinderReservesTheLegacyHelpOptions(t *testing.T) {
	for _, arguments := range [][]string{{"run", "--help"}, {"run", "-h"}} {
		t.Run(strings.Join(arguments, " "), func(t *testing.T) {
			called := 0
			app, output, errors, _ := testApp(t, nil)
			handler := func(context.Context, runcommand.Input) (runcommand.Result, error) {
				called++
				return runcommand.Result{}, nil
			}

			outcome := executeRunBinder(app, arguments, handler, output, errors)

			if outcome.Code != 0 || called != 0 || !strings.Contains(output.String(), "Run package.json scripts") || errors.Len() != 0 {
				t.Fatalf("arguments %q outcome = %#v, calls = %d, stdout = %q, stderr = %q", arguments, outcome, called, output.String(), errors.String())
			}
		})
	}
}

func TestRunBinderAcceptsTheGlobalLogLevelAfterTheLeaf(t *testing.T) {
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	runtime := logging.NewRuntime(logging.Options{Writer: errors})
	var inputs []runcommand.Input
	app, err := New(BuildInfo{Version: "0.0.0-dev"}, Dependencies{
		Out:     output,
		Err:     errors,
		Logging: runtime,
		Run: func(_ context.Context, input runcommand.Input) (runcommand.Result, error) {
			inputs = append(inputs, input)
			return runcommand.Result{}, nil
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	for _, arguments := range [][]string{{"run", "project", "--log-level", "warn"}, {"run", "--log-level=warn", "other"}} {
		outcome := app.Execute(context.Background(), arguments)
		if outcome.Code != 0 || outcome.Err != nil {
			t.Fatalf("arguments %q outcome = %#v, stderr = %q", arguments, outcome, errors.String())
		}
		output.Reset()
		errors.Reset()
	}
	if !reflect.DeepEqual(inputs, []runcommand.Input{{Directory: "project"}, {Directory: "other"}}) {
		t.Fatalf("inputs = %#v", inputs)
	}
	if runtime.Level() != logging.Warn {
		t.Fatalf("log level = %v, want %v", runtime.Level(), logging.Warn)
	}
}

func TestRunBinderRetainsTheFrozenPassthroughRejectionMatrix(t *testing.T) {
	testCases := [][]string{
		{"run", ".", "--flag"},
		{"run", "--flag", "value"},
		{"run", "arg1", "arg2"},
		{"run", "--", "arg1", "arg2"},
	}
	for _, arguments := range testCases {
		t.Run(strings.Join(arguments, " "), func(t *testing.T) {
			called := 0
			app, output, errors, _ := testApp(t, nil)
			handler := func(context.Context, runcommand.Input) (runcommand.Result, error) {
				called++
				return runcommand.Result{}, nil
			}

			outcome := executeRunBinder(app, arguments, handler, output, errors)

			if outcome.Code != 1 || called != 0 || !strings.Contains(errors.String(), "accepts at most 1 arg(s)") {
				t.Fatalf("arguments %q outcome = %#v, calls = %d, stderr = %q", arguments, outcome, called, errors.String())
			}
		})
	}
}

func TestRunBinderMapsTypedChildExitWithoutADiagnostic(t *testing.T) {
	app, output, errors, _ := testApp(t, nil)
	outcome := executeRunBinder(app, []string{"run"}, func(context.Context, runcommand.Input) (runcommand.Result, error) {
		return runcommand.Result{ExitCode: 7}, nil
	}, output, errors)

	if outcome.Code != 7 || outcome.Err != nil || output.Len() != 0 || errors.Len() != 0 {
		t.Fatalf("outcome = %#v, stdout = %q, stderr = %q", outcome, output.String(), errors.String())
	}
}

func TestRunRegistrationExposesTheRealLeaf(t *testing.T) {
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	var inputs []runcommand.Input
	app, err := New(BuildInfo{Version: "0.0.0-dev"}, Dependencies{
		Out: output,
		Err: errors,
		Run: func(_ context.Context, input runcommand.Input) (runcommand.Result, error) {
			inputs = append(inputs, input)
			return runcommand.Result{}, nil
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if outcome := app.Execute(context.Background(), []string{"--help"}); outcome.Code != 0 || !strings.Contains(output.String(), "run") {
		t.Fatalf("root help outcome = %#v, stdout = %q", outcome, output.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"run", "project"}); outcome.Code != 0 || outcome.Err != nil {
		t.Fatalf("run outcome = %#v, stderr = %q", outcome, errors.String())
	}
	if !reflect.DeepEqual(inputs, []runcommand.Input{{Directory: "project"}}) {
		t.Fatalf("inputs = %#v", inputs)
	}
}

func executeRunBinder(app *App, arguments []string, handler RunHandler, output, errors *bytes.Buffer) Outcome {
	root := app.rootCommand()
	root.AddCommand(app.runCommand(handler, func(override string) error {
		return app.configureLogging(override)
	}))
	return app.execute(func() error {
		root.SetArgs(arguments)
		return root.ExecuteContext(context.Background())
	})
}
