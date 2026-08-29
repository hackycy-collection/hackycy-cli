package root

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	heatcommand "github.com/hackycy/hackycy-cli/internal/commands/git/heat"
	"github.com/hackycy/hackycy-cli/internal/logging"
)

func TestGitHeatBindingPassesCompatibilityInputAndGlobalLogLevel(t *testing.T) {
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	runtime := logging.NewRuntime(logging.Options{Writer: errors})
	var inputs []heatcommand.Input
	app, err := newTestApp(BuildInfo{Version: "0.0.0-dev"}, testDependencies{
		Out:     output,
		Err:     errors,
		Logging: runtime,
		GitHeat: func(_ context.Context, input heatcommand.Input) (heatcommand.Result, error) {
			inputs = append(inputs, input)
			return heatcommand.Result{}, nil
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	outcome := app.Execute(context.Background(), []string{
		"--log-level", "warn",
		"git", "heat",
		"--limit", "3oops",
		"--type", "dirs",
		"--sort", "count",
		"--relative-time",
		"--query", "  Api  ",
	})
	if outcome.Code != 0 || outcome.Err != nil {
		t.Fatalf("outcome = %#v, stderr = %q", outcome, errors.String())
	}
	if len(inputs) != 1 {
		t.Fatalf("inputs = %#v", inputs)
	}
	input := inputs[0]
	if input.Limit == nil || *input.Limit != 3 || input.Days != nil {
		t.Fatalf("range input = %#v", input)
	}
	if input.Target != heatcommand.TargetDirectories || input.Sort != heatcommand.SortCount || !input.RelativeTime || input.Query != "  Api  " {
		t.Fatalf("input = %#v", input)
	}
	if runtime.Level() != logging.Warn {
		t.Fatalf("log level = %v, want %v", runtime.Level(), logging.Warn)
	}
}

func TestGitHeatBindingAppliesDefaultsAndSupportsDays(t *testing.T) {
	var inputs []heatcommand.Input
	app, output, errors, _ := newGitHeatTestApp(t, func(_ context.Context, input heatcommand.Input) (heatcommand.Result, error) {
		inputs = append(inputs, input)
		return heatcommand.Result{}, nil
	})
	for _, arguments := range [][]string{
		{"git", "heat"},
		{"git", "heat", "-d", "+2tail", "-t", "files", "-s", "path"},
	} {
		outcome := app.Execute(context.Background(), arguments)
		if outcome.Code != 0 || outcome.Err != nil {
			t.Fatalf("%v outcome = %#v, stderr = %q", arguments, outcome, errors.String())
		}
		output.Reset()
		errors.Reset()
	}
	if len(inputs) != 2 {
		t.Fatalf("inputs = %#v", inputs)
	}
	if inputs[0].Limit != nil || inputs[0].Days != nil || inputs[0].Target != heatcommand.TargetDirectories || inputs[0].Sort != heatcommand.SortPath {
		t.Fatalf("defaults = %#v", inputs[0])
	}
	if inputs[1].Days == nil || *inputs[1].Days != 2 || inputs[1].Target != heatcommand.TargetFiles || inputs[1].Sort != heatcommand.SortPath {
		t.Fatalf("days input = %#v", inputs[1])
	}
}

func TestGitHeatQueryShorthandCoexistsWithGlobalQuiet(t *testing.T) {
	var inputs []heatcommand.Input
	app, _, errors, runtime := newGitHeatTestApp(t, func(_ context.Context, input heatcommand.Input) (heatcommand.Result, error) {
		inputs = append(inputs, input)
		return heatcommand.Result{}, nil
	})

	if outcome := app.Execute(context.Background(), []string{"git", "heat", "-q", "needle"}); outcome.Code != 0 || errors.Len() != 0 || len(inputs) != 1 || inputs[0].Query != "needle" || runtime.Level() != logging.Info {
		t.Fatalf("query shorthand outcome = %#v, inputs = %#v, stderr = %q, level = %v", outcome, inputs, errors.String(), runtime.Level())
	}

	inputs = nil
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"--quiet", "git", "heat", "-q", "needle"}); outcome.Code != 0 || errors.Len() != 0 || len(inputs) != 1 || inputs[0].Query != "needle" || runtime.Level() != logging.Error {
		t.Fatalf("long quiet outcome = %#v, inputs = %#v, stderr = %q, level = %v", outcome, inputs, errors.String(), runtime.Level())
	}
}

func TestGitHeatBindingRejectsInvalidFlagsBeforeHandler(t *testing.T) {
	calls := 0
	app, output, errors, _ := newGitHeatTestApp(t, func(context.Context, heatcommand.Input) (heatcommand.Result, error) {
		calls++
		return heatcommand.Result{}, nil
	})
	testCases := []struct {
		arguments []string
		want      string
	}{
		{arguments: []string{"git", "heat", "-n", "oops"}, want: "error: 'oops' is not a valid integer\n"},
		{arguments: []string{"git", "heat", "-t", "directory"}, want: "error: 'directory' is not a valid report type. Use files or directories.\n"},
		{arguments: []string{"git", "heat", "-s", "date"}, want: "error: 'date' is not a valid sort. Use count or path.\n"},
		{arguments: []string{"git", "heat", "unexpected"}, want: "error: unknown command 'unexpected'; Run 'ycy git heat --help' for usage.\n"},
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

func TestGitHeatGroupExposesOnlyHeat(t *testing.T) {
	app, output, errors, _ := newGitHeatTestApp(t, func(context.Context, heatcommand.Input) (heatcommand.Result, error) {
		return heatcommand.Result{}, nil
	})
	if outcome := app.Execute(context.Background(), []string{"git", "--help"}); outcome.Code != 0 || !strings.Contains(output.String(), "heat") || strings.Contains(output.String(), "pulse") {
		t.Fatalf("git help outcome = %#v, stdout = %q", outcome, output.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"git", "pulse"}); outcome.Code != 1 || errors.String() != "error: unknown command 'pulse'; Run 'ycy git --help' for usage.\n" {
		t.Fatalf("absent sibling outcome = %#v, stderr = %q", outcome, errors.String())
	}
}

func TestParseHeatIntegerPreservesPermissiveDecimalPrefixes(t *testing.T) {
	testCases := []struct {
		value string
		want  int
	}{
		{value: "3oops", want: 3},
		{value: "  +12tail", want: 12},
		{value: "-0x1", want: 0},
	}
	for _, testCase := range testCases {
		got, err := parseHeatInteger(testCase.value)
		if err != nil || got != testCase.want {
			t.Fatalf("parseHeatInteger(%q) = (%d, %v), want %d", testCase.value, got, err, testCase.want)
		}
	}
	for _, value := range []string{"", " ", "+", "-", "Infinity", "999999999999999999999999999999"} {
		if _, err := parseHeatInteger(value); err == nil {
			t.Fatalf("parseHeatInteger(%q) error = nil", value)
		}
	}
}

func newGitHeatTestApp(t *testing.T, handler GitHeatHandler) (*App, *bytes.Buffer, *bytes.Buffer, *logging.Runtime) {
	t.Helper()
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	runtime := logging.NewRuntime(logging.Options{Writer: errors})
	app, err := newTestApp(BuildInfo{Version: "0.0.0-dev"}, testDependencies{
		Out:     output,
		Err:     errors,
		Logging: runtime,
		GitHeat: handler,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return app, output, errors, runtime
}

func TestGitHeatBindingRetainsBothRangesForTheModuleToReject(t *testing.T) {
	var input heatcommand.Input
	app, _, errors, _ := newGitHeatTestApp(t, func(_ context.Context, value heatcommand.Input) (heatcommand.Result, error) {
		input = value
		return heatcommand.Result{}, nil
	})
	outcome := app.Execute(context.Background(), []string{"git", "heat", "-n", "1", "-d", "2"})
	if outcome.Code != 0 || errors.Len() != 0 {
		t.Fatalf("outcome = %#v, stderr = %q", outcome, errors.String())
	}
	if input.Limit == nil || input.Days == nil || !reflect.DeepEqual([]int{*input.Limit, *input.Days}, []int{1, 2}) {
		t.Fatalf("input = %#v", input)
	}
}

func TestGitHeatBindingPreservesTypedSignalExitWithoutADiagnostic(t *testing.T) {
	app, output, errors, _ := newGitHeatTestApp(t, func(context.Context, heatcommand.Input) (heatcommand.Result, error) {
		return heatcommand.Result{}, gitHeatExitError{code: 143}
	})
	outcome := app.Execute(context.Background(), []string{"git", "heat"})
	if outcome.Code != 143 || outcome.Err != nil || output.Len() != 0 || errors.Len() != 0 {
		t.Fatalf("outcome = %#v, stdout = %q, stderr = %q", outcome, output.String(), errors.String())
	}
}

type gitHeatExitError struct {
	code int
}

func (err gitHeatExitError) Error() string {
	return "git heat signal outcome"
}

func (err gitHeatExitError) ExitCode() int {
	return err.code
}
