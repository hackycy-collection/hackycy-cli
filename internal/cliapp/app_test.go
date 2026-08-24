package cliapp

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	configfork "github.com/hackycy/hackycy-cli/internal/commands/config/fork"
	"github.com/hackycy/hackycy-cli/internal/commands/exportenv"
	"github.com/hackycy/hackycy-cli/internal/logging"
)

func TestGlobalSurface(t *testing.T) {
	app, output, errors, runtime := testApp(t, nil)

	if outcome := app.Execute(context.Background(), nil); outcome.Code != 1 || !strings.Contains(output.String(), "Usage:") || errors.Len() != 0 {
		t.Fatalf("no-argument outcome = %#v, stdout = %q, stderr = %q", outcome, output.String(), errors.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"--help"}); outcome.Code != 0 || !strings.Contains(output.String(), "Usage:") || errors.Len() != 0 {
		t.Fatalf("help outcome = %#v, stdout = %q, stderr = %q", outcome, output.String(), errors.String())
	}
	output.Reset()
	if outcome := app.Execute(context.Background(), []string{"-V"}); outcome.Code != 0 || output.String() != "0.0.0-dev\n" {
		t.Fatalf("version outcome = %#v, stdout = %q", outcome, output.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"unknown"}); outcome.Code != 1 || errors.String() != "error: unknown command 'unknown'\n" {
		t.Fatalf("unknown outcome = %#v, stderr = %q", outcome, errors.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"--log-level", "warn"}); outcome.Code != 1 || runtime.Level() != logging.Warn {
		t.Fatalf("log-level outcome = %#v, level = %v", outcome, runtime.Level())
	}
}

func TestGlobalLogPrecedenceAndVersionBypass(t *testing.T) {
	app, output, errors, runtime := testApp(t, map[string]string{"YCY_LOG_LEVEL": "debug"})
	if outcome := app.Execute(context.Background(), nil); outcome.Code != 1 || runtime.Level() != logging.Debug {
		t.Fatalf("environment outcome = %#v, level = %v", outcome, runtime.Level())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"--version", "--log-level", "invalid"}); outcome.Code != 0 || output.String() != "0.0.0-dev\n" || errors.Len() != 0 {
		t.Fatalf("version bypass outcome = %#v, stdout = %q, stderr = %q", outcome, output.String(), errors.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"--log-level", "invalid"}); outcome.Code != 1 || !strings.Contains(errors.String(), "invalid log level") {
		t.Fatalf("invalid log outcome = %#v, stderr = %q", outcome, errors.String())
	}
}

func TestAppconfigFoundationDoesNotExposeACommand(t *testing.T) {
	app, output, errors, _ := testApp(t, nil)

	if outcome := app.Execute(context.Background(), []string{"--help"}); outcome.Code != 0 || strings.Contains(output.String(), "config") {
		t.Fatalf("help outcome = %#v, stdout = %q", outcome, output.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"config"}); outcome.Code != 1 || errors.String() != "error: unknown command 'config'\n" {
		t.Fatalf("config outcome = %#v, stderr = %q", outcome, errors.String())
	}
}

func TestFSFoundationDoesNotExposeACommand(t *testing.T) {
	app, output, errors, _ := testApp(t, nil)

	if outcome := app.Execute(context.Background(), []string{"--help"}); outcome.Code != 0 || strings.Contains(output.String(), "\n  fs") {
		t.Fatalf("help outcome = %#v, stdout = %q", outcome, output.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"fs"}); outcome.Code != 1 || errors.String() != "error: unknown command 'fs'\n" {
		t.Fatalf("fs outcome = %#v, stderr = %q", outcome, errors.String())
	}
}

func TestPanicMappingRedactsAndAddsDebugStack(t *testing.T) {
	app, output, errors, _ := testApp(t, map[string]string{"DEBUG": "1"})
	outcome := app.execute(func() error {
		panic("token=not-for-output")
	})
	if outcome.Code != 1 || output.String() != "\n" || strings.Contains(errors.String(), "not-for-output") || !strings.Contains(errors.String(), "token=[REDACTED]") || !strings.Contains(errors.String(), "goroutine") {
		t.Fatalf("panic outcome = %#v, stdout = %q, stderr = %q", outcome, output.String(), errors.String())
	}
}

func TestExportEnvBindingPassesTypedInputAndGlobalLogLevel(t *testing.T) {
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	runtime := logging.NewRuntime(logging.Options{Writer: errors})
	var inputs []exportenv.Input
	app, err := New(BuildInfo{Version: "0.0.0-dev"}, Dependencies{
		Out:     output,
		Err:     errors,
		Logging: runtime,
		ExportEnv: func(_ context.Context, input exportenv.Input) (exportenv.Result, error) {
			inputs = append(inputs, input)
			return exportenv.Result{}, nil
		},
	})
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}

	outcome := app.Execute(context.Background(), []string{
		"--log-level", "warn",
		"export", "env", "project",
		"-e", "production",
		"--merge",
		"-o", "output.json",
	})

	if outcome.Code != 0 || outcome.Err != nil {
		t.Fatalf("outcome = %#v, stderr = %q", outcome, errors.String())
	}
	want := []exportenv.Input{{
		Directory:   "project",
		Environment: "production",
		Merge:       true,
		Output:      "output.json",
	}}
	if !reflect.DeepEqual(inputs, want) {
		t.Fatalf("inputs = %#v, want %#v", inputs, want)
	}
	if runtime.Level() != logging.Warn {
		t.Fatalf("log level = %v, want %v", runtime.Level(), logging.Warn)
	}
}

func TestConfigForkListBindingPassesTypedInputAndExposesNoSibling(t *testing.T) {
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	runtime := logging.NewRuntime(logging.Options{Writer: errors})
	var inputs []configfork.Input
	app, err := New(BuildInfo{Version: "0.0.0-dev"}, Dependencies{
		Out:     output,
		Err:     errors,
		Logging: runtime,
		ConfigForkList: func(_ context.Context, input configfork.Input) (configfork.Result, error) {
			inputs = append(inputs, input)
			return configfork.Result{}, nil
		},
	})
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}

	outcome := app.Execute(context.Background(), []string{"--log-level", "warn", "config", "fork", "list"})
	if outcome.Code != 0 || outcome.Err != nil {
		t.Fatalf("list outcome = %#v, stderr = %q", outcome, errors.String())
	}
	if !reflect.DeepEqual(inputs, []configfork.Input{{}}) {
		t.Fatalf("inputs = %#v", inputs)
	}
	if runtime.Level() != logging.Warn {
		t.Fatalf("log level = %v, want %v", runtime.Level(), logging.Warn)
	}

	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"config", "fork", "--help"}); outcome.Code != 0 || !strings.Contains(output.String(), "list") || strings.Contains(output.String(), "add") || strings.Contains(output.String(), "remove") {
		t.Fatalf("fork help outcome = %#v, stdout = %q, stderr = %q", outcome, output.String(), errors.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"config", "fork", "add"}); outcome.Code != 1 || errors.String() != "error: unknown command 'add'\n" {
		t.Fatalf("absent sibling outcome = %#v, stderr = %q", outcome, errors.String())
	}
}

func testApp(t *testing.T, environment map[string]string) (*App, *bytes.Buffer, *bytes.Buffer, *logging.Runtime) {
	t.Helper()
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	runtime := logging.NewRuntime(logging.Options{Writer: errors})
	app, err := New(BuildInfo{Version: "0.0.0-dev"}, Dependencies{
		Out: output,
		Err: errors,
		Environment: func(key string) string {
			return environment[key]
		},
		Logging: runtime,
	})
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}
	return app, output, errors, runtime
}
