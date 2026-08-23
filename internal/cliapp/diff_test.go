package cliapp

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/hackycy/hackycy-cli/internal/commands/diff"
)

func TestDiffBindingPassesTypedInputAndLegacyDefaults(t *testing.T) {
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	var inputs []diff.Input
	app, err := New(BuildInfo{Version: "0.0.0-dev"}, Dependencies{
		Out: output,
		Err: errors,
		Diff: func(_ context.Context, input diff.Input) (diff.Result, error) {
			inputs = append(inputs, input)
			return diff.Result{}, nil
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	outcome := app.Execute(context.Background(), []string{
		"diff", "baseline", "-x", "first", "--port", "00007", "target", "--exclude=second", "--public", "--no-gitignore",
	})
	if outcome.Code != 0 || outcome.Err != nil {
		t.Fatalf("configured diff outcome = %#v, stderr = %q", outcome, errors.String())
	}

	outcome = app.Execute(context.Background(), []string{"diff", "before", "after"})
	if outcome.Code != 0 || outcome.Err != nil {
		t.Fatalf("default diff outcome = %#v, stderr = %q", outcome, errors.String())
	}

	want := []diff.Input{
		{
			BaselineDirectory: "baseline",
			TargetDirectory:   "target",
			Port:              7,
			Public:            true,
			Exclusions:        []string{"first", "second"},
			NoGitIgnore:       true,
		},
		{
			BaselineDirectory: "before",
			TargetDirectory:   "after",
			Port:              1205,
			Exclusions:        []string{},
		},
	}
	if !reflect.DeepEqual(inputs, want) {
		t.Fatalf("inputs = %#v, want %#v", inputs, want)
	}
}

func TestDiffBindingRejectsInvalidPortsAndOperandCounts(t *testing.T) {
	invalidPorts := []struct {
		value string
		want  string
	}{
		{value: "-1", want: "'-1' is not a valid port"},
		{value: "+1", want: "'+1' is not a valid port"},
		{value: " 1", want: "' 1' is not a valid port"},
		{value: "1.0", want: "'1.0' is not a valid port"},
		{value: "0x50", want: "'0x50' is not a valid port"},
		{value: "\uff11", want: "'\uff11' is not a valid port"},
		{value: "65536", want: "Port must be between 0 and 65535"},
	}
	for _, testCase := range invalidPorts {
		t.Run(testCase.value, func(t *testing.T) {
			app, output, errors, calls := diffTestApp(t)
			outcome := app.Execute(context.Background(), []string{"diff", "--port=" + testCase.value, "baseline", "target"})
			if outcome.Code != 1 || *calls != 0 || errors.String() != "error: "+testCase.want+"\n" {
				t.Fatalf("port %q outcome = %#v, calls = %d, stdout = %q, stderr = %q", testCase.value, outcome, *calls, output.String(), errors.String())
			}
		})
	}

	for _, arguments := range [][]string{{"diff", "baseline"}, {"diff", "baseline", "target", "extra"}} {
		t.Run(strings.Join(arguments, " "), func(t *testing.T) {
			app, output, errors, calls := diffTestApp(t)
			outcome := app.Execute(context.Background(), arguments)
			if outcome.Code != 1 || *calls != 0 || !strings.Contains(errors.String(), "accepts 2 arg(s)") {
				t.Fatalf("arguments %q outcome = %#v, calls = %d, stdout = %q, stderr = %q", arguments, outcome, *calls, output.String(), errors.String())
			}
		})
	}
}

func TestDiffBindingExposesOnlyAnInjectedHandlerAndItsHelp(t *testing.T) {
	app, output, errors, _ := testApp(t, nil)
	if outcome := app.Execute(context.Background(), []string{"--help"}); outcome.Code != 0 || strings.Contains(output.String(), "diff") {
		t.Fatalf("unregistered root help outcome = %#v, stdout = %q", outcome, output.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"diff"}); outcome.Code != 1 || errors.String() != "error: unknown command 'diff'\n" {
		t.Fatalf("unregistered diff outcome = %#v, stderr = %q", outcome, errors.String())
	}

	app, output, errors, _ = diffTestApp(t)
	if outcome := app.Execute(context.Background(), []string{"diff", "--help"}); outcome.Code != 0 || !strings.Contains(output.String(), "Compare two directories in a browser") || !strings.Contains(output.String(), "--port") || !strings.Contains(output.String(), "--public") || !strings.Contains(output.String(), "--exclude") || !strings.Contains(output.String(), "--no-gitignore") || strings.Contains(output.String(), "--address") {
		t.Fatalf("registered diff help outcome = %#v, stdout = %q, stderr = %q", outcome, output.String(), errors.String())
	}
}

func diffTestApp(t *testing.T) (*App, *bytes.Buffer, *bytes.Buffer, *int) {
	t.Helper()
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	calls := 0
	app, err := New(BuildInfo{Version: "0.0.0-dev"}, Dependencies{
		Out: output,
		Err: errors,
		Diff: func(context.Context, diff.Input) (diff.Result, error) {
			calls++
			return diff.Result{}, nil
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return app, output, errors, &calls
}
