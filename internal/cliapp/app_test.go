package cliapp

import (
	"bytes"
	"context"
	"strings"
	"testing"

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

func TestPanicMappingRedactsAndAddsDebugStack(t *testing.T) {
	app, output, errors, _ := testApp(t, map[string]string{"DEBUG": "1"})
	outcome := app.execute(func() error {
		panic("token=not-for-output")
	})
	if outcome.Code != 1 || output.String() != "\n" || strings.Contains(errors.String(), "not-for-output") || !strings.Contains(errors.String(), "token=[REDACTED]") || !strings.Contains(errors.String(), "goroutine") {
		t.Fatalf("panic outcome = %#v, stdout = %q, stderr = %q", outcome, output.String(), errors.String())
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
