package cliapp

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/hackycy/hackycy-cli/internal/commands/rm"
	"github.com/hackycy/hackycy-cli/internal/logging"
)

func TestRmBindingPassesTypedInputAndLegacyDepthPrefixes(t *testing.T) {
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	runtime := logging.NewRuntime(logging.Options{Writer: errors})
	var inputs []rm.Input
	app, err := New(BuildInfo{Version: "0.0.0-dev"}, Dependencies{
		Out:     output,
		Err:     errors,
		Logging: runtime,
		RM: func(_ context.Context, input rm.Input) (rm.Result, error) {
			inputs = append(inputs, input)
			return rm.Result{}, nil
		},
	})
	if err != nil {
		t.Fatalf("New() returned an error: %v", err)
	}

	outcome := app.Execute(context.Background(), []string{"--log-level", "warn", "rm", "--force", "--depth", "3ignored", "one", "two"})

	if outcome.Code != 0 || outcome.Err != nil {
		t.Fatalf("rm outcome = %#v, stderr = %q", outcome, errors.String())
	}
	depth := 3
	want := []rm.Input{{Paths: []string{"one", "two"}, Force: true, Depth: &depth}}
	if !reflect.DeepEqual(inputs, want) {
		t.Fatalf("rm inputs = %#v, want %#v", inputs, want)
	}
	if runtime.Level() != logging.Warn {
		t.Fatalf("log level = %v, want %v", runtime.Level(), logging.Warn)
	}

	output.Reset()
	errors.Reset()
	outcome = app.Execute(context.Background(), []string{"rm", "--depth=-2"})
	if outcome.Code != 0 || len(inputs) != 2 || inputs[1].Depth == nil || *inputs[1].Depth != -2 {
		t.Fatalf("negative depth outcome = %#v, inputs = %#v", outcome, inputs)
	}
}

func TestRmBindingRejectsInvalidDepthAndExposesTheRealLeaf(t *testing.T) {
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	called := 0
	app, err := New(BuildInfo{Version: "0.0.0-dev"}, Dependencies{
		Out: output,
		Err: errors,
		RM: func(context.Context, rm.Input) (rm.Result, error) {
			called++
			return rm.Result{}, nil
		},
	})
	if err != nil {
		t.Fatalf("New() returned an error: %v", err)
	}

	outcome := app.Execute(context.Background(), []string{"rm", "--depth", "not-a-number"})
	if outcome.Code != 1 || called != 0 || errors.String() != "error: 'not-a-number' is not a valid integer\n" {
		t.Fatalf("invalid depth outcome = %#v, calls = %d, stderr = %q", outcome, called, errors.String())
	}

	output.Reset()
	errors.Reset()
	outcome = app.Execute(context.Background(), []string{"--help"})
	if outcome.Code != 0 || !strings.Contains(output.String(), "rm") {
		t.Fatalf("root help outcome = %#v, stdout = %q", outcome, output.String())
	}
	output.Reset()
	errors.Reset()
	outcome = app.Execute(context.Background(), []string{"rm", "--help"})
	if outcome.Code != 0 || !strings.Contains(output.String(), "--force") || !strings.Contains(output.String(), "--depth") {
		t.Fatalf("rm help outcome = %#v, stdout = %q", outcome, output.String())
	}
}
