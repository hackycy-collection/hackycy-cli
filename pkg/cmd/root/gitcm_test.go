package root

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	cmcommand "github.com/hackycy/hackycy-cli/internal/commands/git/cm"
	"github.com/hackycy/hackycy-cli/internal/logging"
)

func TestGitCMBindingPassesLegacyFlagMatrix(t *testing.T) {
	var inputs []cmcommand.Input
	app, output, errors, runtime := newGitCMTestApp(t, func(_ context.Context, input cmcommand.Input) (cmcommand.Result, error) {
		inputs = append(inputs, input)
		return cmcommand.Result{}, nil
	})

	for index, arguments := range [][]string{
		{"git", "cm"},
		{"--log-level", "warn", "git", "cm", "--profile", "work", "--timeout-ms", "0x3e8", "-l", "zh", "-S", "-s", "-a", "-d", "-b"},
		{"git", "cm", "--push"},
		{"git", "cm", "--stage-push=upstream"},
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

	if len(inputs) != 4 {
		t.Fatalf("inputs = %#v", inputs)
	}
	if got := inputs[0]; got != (cmcommand.Input{Language: "en"}) {
		t.Fatalf("default input = %#v", got)
	}
	if got := inputs[1]; got.Profile != "work" || got.TimeoutMS == nil || *got.TimeoutMS != 1000 || got.Language != "zh" || !got.Staged || !got.Stage || !got.StageAll || !got.DryRun || !got.Body {
		t.Fatalf("full input = %#v", got)
	}
	if got := inputs[2]; got.Push == nil || *got.Push != "origin" || got.StagePush != nil {
		t.Fatalf("bare push input = %#v", got)
	}
	if got := inputs[3]; got.Push != nil || got.StagePush == nil || *got.StagePush != "upstream" {
		t.Fatalf("stage push input = %#v", got)
	}
}

func TestGitCMBindingNormalizesCommanderOptionalRemoteForms(t *testing.T) {
	var inputs []cmcommand.Input
	app, output, errors, _ := newGitCMTestApp(t, func(_ context.Context, input cmcommand.Input) (cmcommand.Result, error) {
		inputs = append(inputs, input)
		return cmcommand.Result{}, nil
	})

	for _, arguments := range [][]string{
		{"git", "cm", "--push", "upstream"},
		{"git", "cm", "-p", "upstream"},
		{"git", "cm", "-pupstream"},
		{"git", "cm", "--stage-push", "publish"},
		{"git", "cm", "-cpublish"},
		{"git", "cm", "-p=upstream"},
		{"git", "cm", "--push", "-d"},
	} {
		outcome := app.Execute(context.Background(), arguments)
		if outcome.Code != 0 || outcome.Err != nil {
			t.Fatalf("%v outcome = %#v, stderr = %q", arguments, outcome, errors.String())
		}
		output.Reset()
		errors.Reset()
	}

	if got, want := inputs, []cmcommand.Input{
		{Language: "en", Push: cmInputString("upstream")},
		{Language: "en", Push: cmInputString("upstream")},
		{Language: "en", Push: cmInputString("upstream")},
		{Language: "en", StagePush: cmInputString("publish")},
		{Language: "en", StagePush: cmInputString("publish")},
		{Language: "en", Push: cmInputString("=upstream")},
		{Language: "en", Push: cmInputString("origin"), DryRun: true},
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("inputs = %#v, want %#v", got, want)
	}
}

func TestGitCMBindingRejectsInvalidTimeoutBeforeHandler(t *testing.T) {
	calls := 0
	app, output, errors, _ := newGitCMTestApp(t, func(context.Context, cmcommand.Input) (cmcommand.Result, error) {
		calls++
		return cmcommand.Result{}, nil
	})
	for _, value := range []string{"999", "1.5", "not-a-number", "9007199254740992"} {
		outcome := app.Execute(context.Background(), []string{"git", "cm", "--timeout-ms", value})
		want := "error: '" + value + "' is not a valid timeout in milliseconds. Use an integer greater than or equal to 1000.\n"
		if outcome.Code != 1 || errors.String() != want {
			t.Fatalf("%q outcome = %#v, stderr = %q, want %q", value, outcome, errors.String(), want)
		}
		output.Reset()
		errors.Reset()
	}
	if calls != 0 {
		t.Fatalf("handler calls = %d", calls)
	}
}

func TestParseCMTimeoutMSMatchesStrictJavaScriptNumberSemantics(t *testing.T) {
	for _, testCase := range []struct {
		value string
		want  float64
	}{
		{value: "1000", want: 1000},
		{value: "1e3", want: 1000},
		{value: "0b1111101000", want: 1000},
		{value: "0o1750", want: 1000},
		{value: "\uFEFF1001", want: 1001},
		{value: "9007199254740991", want: 9007199254740991},
	} {
		got, err := parseCMTimeoutMS(testCase.value)
		if err != nil || got != testCase.want {
			t.Fatalf("parseCMTimeoutMS(%q) = (%v, %v), want (%v, nil)", testCase.value, got, err, testCase.want)
		}
	}
	for _, value := range []string{"", "999", "1.5", "Infinity", "NaN", "1000suffix", "9007199254740992"} {
		if _, err := parseCMTimeoutMS(value); err == nil {
			t.Fatalf("parseCMTimeoutMS(%q) error = nil", value)
		}
	}
	if _, err := parseCMTimeoutMS("1000.0"); err != nil {
		t.Fatalf("parseCMTimeoutMS decimal integer error = %v", err)
	}
	if _, err := parseCMTimeoutMS("-1000"); err == nil {
		t.Fatalf("parseCMTimeoutMS negative result = %v", err)
	}
}

func TestNormalizeGitCMArgumentsLeavesOtherCommandArgumentsUntouched(t *testing.T) {
	arguments := []string{"git", "pulse", "--push", "upstream"}
	if got := normalizeGitCMArguments(arguments); !reflect.DeepEqual(got, arguments) {
		t.Fatalf("normalized arguments = %#v, want %#v", got, arguments)
	}
	arguments = []string{"git", "cm", "--", "--push", "upstream"}
	if got := normalizeGitCMArguments(arguments); !reflect.DeepEqual(got, arguments) {
		t.Fatalf("normalized arguments after delimiter = %#v, want %#v", got, arguments)
	}
}

func TestGitCMGroupIsAbsentWithoutTheProductionHandler(t *testing.T) {
	app, output, errors, _ := testApp(t, nil)
	if outcome := app.Execute(context.Background(), []string{"git", "cm"}); outcome.Code != 1 || errors.String() != "error: unknown command 'git'; did you mean 'zip'? Run 'ycy zip --help' for usage.\n" || output.Len() != 0 {
		t.Fatalf("outcome = %#v, stdout = %q, stderr = %q", outcome, output.String(), errors.String())
	}
}

func newGitCMTestApp(t *testing.T, handler GitCMHandler) (*App, *bytes.Buffer, *bytes.Buffer, *logging.Runtime) {
	t.Helper()
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	runtime := logging.NewRuntime(logging.Options{Writer: errors})
	app, err := newTestApp(BuildInfo{Version: "0.0.0-dev"}, testDependencies{
		Out:     output,
		Err:     errors,
		Logging: runtime,
		GitCM:   handler,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return app, output, errors, runtime
}

func cmInputString(value string) *string {
	return &value
}

func TestGitCMBindingRetainsTypedSignalExitWithoutADiagnostic(t *testing.T) {
	app, output, errors, _ := newGitCMTestApp(t, func(context.Context, cmcommand.Input) (cmcommand.Result, error) {
		return cmcommand.Result{}, gitCMExitError{code: 143}
	})
	outcome := app.Execute(context.Background(), []string{"git", "cm"})
	if outcome.Code != 143 || outcome.Err != nil || output.Len() != 0 || errors.Len() != 0 || strings.Contains(errors.String(), "git cm") {
		t.Fatalf("outcome = %#v, stdout = %q, stderr = %q", outcome, output.String(), errors.String())
	}
}

type gitCMExitError struct {
	code int
}

func (err gitCMExitError) Error() string {
	return "git cm signal outcome"
}

func (err gitCMExitError) ExitCode() int {
	return err.code
}
