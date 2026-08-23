package cliapp

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	pulsecommand "github.com/hackycy/hackycy-cli/internal/commands/git/pulse"
	"github.com/hackycy/hackycy-cli/internal/logging"
)

func TestGitPulseBindingPassesOptionalDirectoryAndLegacyDays(t *testing.T) {
	var inputs []pulsecommand.Input
	app, output, errors, runtime := newGitPulseTestApp(t, func(_ context.Context, input pulsecommand.Input) (pulsecommand.Result, error) {
		inputs = append(inputs, input)
		return pulsecommand.Result{}, nil
	})

	for index, arguments := range [][]string{
		{"git", "pulse"},
		{"--log-level", "warn", "git", "pulse", "workspace", "--days", "3oops"},
		{"git", "pulse", "--days=-1tail"},
		{"git", "pulse", "--days", "0"},
	} {
		outcome := app.Execute(context.Background(), arguments)
		if outcome.Code != 0 || outcome.Err != nil {
			t.Fatalf("%v outcome = %#v, stderr = %q", arguments, outcome, errors.String())
		}
		if index == 1 && runtime.Level() != logging.Warn {
			t.Fatalf("log level = %v, want %v", runtime.Level(), logging.Warn)
		}
		output.Reset()
		errors.Reset()
	}

	want := []pulsecommand.Input{
		{},
		{Directory: "workspace", Days: pulseInt(3)},
		{Days: pulseInt(-1)},
		{Days: pulseInt(0)},
	}
	if !reflect.DeepEqual(inputs, want) {
		t.Fatalf("inputs = %#v, want %#v", inputs, want)
	}
}

func TestGitPulseBindingRejectsInvalidArgumentsBeforeHandler(t *testing.T) {
	calls := 0
	app, output, errors, _ := newGitPulseTestApp(t, func(context.Context, pulsecommand.Input) (pulsecommand.Result, error) {
		calls++
		return pulsecommand.Result{}, nil
	})
	testCases := []struct {
		arguments []string
		want      string
	}{
		{arguments: []string{"git", "pulse", "--days", "oops"}, want: "error: 'oops' is not a valid integer\n"},
		{arguments: []string{"git", "pulse", "one", "two"}, want: "error: accepts at most 1 arg(s), received 2\n"},
		{arguments: []string{"git", "pulse", "--unknown"}, want: "error: unknown flag: --unknown\n"},
	}
	for _, testCase := range testCases {
		outcome := app.Execute(context.Background(), testCase.arguments)
		if outcome.Code != 1 || errors.String() != testCase.want {
			t.Fatalf("%v outcome = %#v, stderr = %q, want %q", testCase.arguments, outcome, errors.String(), testCase.want)
		}
		output.Reset()
		errors.Reset()
	}
	if calls != 0 {
		t.Fatalf("handler calls = %d", calls)
	}
}

func TestGitPulseGroupExposesOnlyPulseWhenHeatIsAbsent(t *testing.T) {
	app, output, errors, _ := newGitPulseTestApp(t, func(context.Context, pulsecommand.Input) (pulsecommand.Result, error) {
		return pulsecommand.Result{}, nil
	})
	if outcome := app.Execute(context.Background(), []string{"git", "--help"}); outcome.Code != 0 || !strings.Contains(output.String(), "pulse") || strings.Contains(output.String(), "heat") {
		t.Fatalf("git help outcome = %#v, stdout = %q", outcome, output.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"git", "heat"}); outcome.Code != 1 || errors.String() != "error: unknown command 'heat'\n" {
		t.Fatalf("absent sibling outcome = %#v, stderr = %q", outcome, errors.String())
	}
}

func TestGitPulseBindingPreservesTypedSignalExitWithoutADiagnostic(t *testing.T) {
	app, output, errors, _ := newGitPulseTestApp(t, func(context.Context, pulsecommand.Input) (pulsecommand.Result, error) {
		return pulsecommand.Result{}, gitPulseExitError{code: 143}
	})
	outcome := app.Execute(context.Background(), []string{"git", "pulse"})
	if outcome.Code != 143 || outcome.Err != nil || output.Len() != 0 || errors.Len() != 0 {
		t.Fatalf("outcome = %#v, stdout = %q, stderr = %q", outcome, output.String(), errors.String())
	}
}

func TestParsePulseIntegerPreservesPermissiveDecimalPrefixes(t *testing.T) {
	testCases := []struct {
		value string
		want  int
	}{
		{value: "3oops", want: 3},
		{value: "  +12tail", want: 12},
		{value: "-0x1", want: 0},
	}
	for _, testCase := range testCases {
		got, err := parsePulseInteger(testCase.value)
		if err != nil || got != testCase.want {
			t.Fatalf("parsePulseInteger(%q) = (%d, %v), want %d", testCase.value, got, err, testCase.want)
		}
	}
	for _, value := range []string{"", " ", "+", "-", "Infinity", "999999999999999999999999999999999"} {
		if _, err := parsePulseInteger(value); err == nil {
			t.Fatalf("parsePulseInteger(%q) error = nil", value)
		}
	}
}

func newGitPulseTestApp(t *testing.T, handler GitPulseHandler) (*App, *bytes.Buffer, *bytes.Buffer, *logging.Runtime) {
	t.Helper()
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	runtime := logging.NewRuntime(logging.Options{Writer: errors})
	app, err := New(BuildInfo{Version: "0.0.0-dev"}, Dependencies{
		Out:      output,
		Err:      errors,
		Logging:  runtime,
		GitPulse: handler,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return app, output, errors, runtime
}

func pulseInt(value int) *int {
	return &value
}

type gitPulseExitError struct {
	code int
}

func (err gitPulseExitError) Error() string {
	return "git pulse signal outcome"
}

func (err gitPulseExitError) ExitCode() int {
	return err.code
}
