package root

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	zipcommand "github.com/hackycy/hackycy-cli/internal/commands/zip"
	"github.com/hackycy/hackycy-cli/internal/logging"
)

func TestZipBindingPassesLegacyFlagsAndDefaults(t *testing.T) {
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	runtime := logging.NewRuntime(logging.Options{Writer: errors})
	var inputs []zipcommand.Input
	app, err := newTestApp(BuildInfo{Version: "0.0.0-dev"}, testDependencies{
		Out:     output,
		Err:     errors,
		Logging: runtime,
		ZIP: func(_ context.Context, input zipcommand.Input) (zipcommand.Result, error) {
			inputs = append(inputs, input)
			return zipcommand.Result{}, nil
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	outcome := app.Execute(context.Background(), []string{"--log-level", "warn", "zip", "project", "-w", "-d", "../unsafe"})
	if outcome.Code != 0 || outcome.Err != nil {
		t.Fatalf("zip outcome = %#v, stderr = %q", outcome, errors.String())
	}
	want := []zipcommand.Input{{Directory: "project", Open: false, WithDir: "../unsafe"}}
	if !reflect.DeepEqual(inputs, want) {
		t.Fatalf("inputs = %#v, want %#v", inputs, want)
	}
	if runtime.Level() != logging.Warn {
		t.Fatalf("log level = %v, want %v", runtime.Level(), logging.Warn)
	}

	output.Reset()
	errors.Reset()
	outcome = app.Execute(context.Background(), []string{"zip"})
	if outcome.Code != 0 || len(inputs) != 2 || !reflect.DeepEqual(inputs[1], zipcommand.Input{Open: true}) {
		t.Fatalf("default zip outcome = %#v, inputs = %#v", outcome, inputs)
	}
}

func TestZipBindingExposesOnlyTheRealLeafAndRejectsExtraOperands(t *testing.T) {
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	called := 0
	app, err := newTestApp(BuildInfo{Version: "0.0.0-dev"}, testDependencies{
		Out: output,
		Err: errors,
		ZIP: func(context.Context, zipcommand.Input) (zipcommand.Result, error) {
			called++
			return zipcommand.Result{}, nil
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	outcome := app.Execute(context.Background(), []string{"zip", "one", "two"})
	if outcome.Code != 1 || called != 0 || !strings.Contains(errors.String(), "accepts at most 1 arg(s)") {
		t.Fatalf("extra operands outcome = %#v, calls = %d, stderr = %q", outcome, called, errors.String())
	}
	output.Reset()
	errors.Reset()
	outcome = app.Execute(context.Background(), []string{"zip", "--help"})
	if outcome.Code != 0 || !strings.Contains(output.String(), "Zip a directory into a zip file") || !strings.Contains(output.String(), "--without-open") || !strings.Contains(output.String(), "--with-dir") {
		t.Fatalf("zip help outcome = %#v, stdout = %q", outcome, output.String())
	}
}
