package root

import (
	"bytes"
	"context"
	stderrors "errors"
	"strings"
	"testing"

	githeat "github.com/hackycy/hackycy-cli/internal/commands/git/heat"
	"github.com/hackycy/hackycy-cli/internal/logging"
	"github.com/hackycy/hackycy-cli/internal/terminal"
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
	if outcome := app.Execute(context.Background(), []string{"unknown"}); outcome.Code != 1 || errors.String() != "error: unknown command 'unknown'; Run 'ycy --help' for usage.\n" {
		t.Fatalf("unknown outcome = %#v, stderr = %q", outcome, errors.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"--log-level", "warn"}); outcome.Code != 1 || runtime.Level() != logging.Info {
		t.Fatalf("log-level outcome = %#v, level = %v", outcome, runtime.Level())
	}
}

func TestGlobalLogPrecedenceAndVersionBypass(t *testing.T) {
	app, output, errors, runtime := testApp(t, map[string]string{"YCY_LOG_LEVEL": "debug"})
	if outcome := app.Execute(context.Background(), nil); outcome.Code != 1 || runtime.Level() != logging.Info {
		t.Fatalf("environment outcome = %#v, level = %v", outcome, runtime.Level())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"--version", "--log-level", "invalid"}); outcome.Code != 0 || output.String() != "0.0.0-dev\n" || errors.Len() != 0 {
		t.Fatalf("version bypass outcome = %#v, stdout = %q, stderr = %q", outcome, output.String(), errors.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"--log-level", "invalid"}); outcome.Code != 1 || errors.Len() != 0 || !strings.Contains(output.String(), "Usage:") {
		t.Fatalf("root discovery bypass outcome = %#v, stdout = %q, stderr = %q", outcome, output.String(), errors.String())
	}
}

func TestDiagnosticConfigurationRunsBeforeEffectsAndBypassesDiscovery(t *testing.T) {
	tests := []struct {
		name        string
		environment map[string]string
		arguments   []string
		wantCode    int
		wantLevel   logging.Level
		wantFormat  logging.RecordFormat
		wantError   string
		wantCall    bool
	}{
		{
			name:        "environment config applies to executing command",
			environment: map[string]string{"YCY_LOG_LEVEL": "warn", "YCY_LOG_FORMAT": "json"},
			arguments:   []string{"git", "heat"},
			wantLevel:   logging.Warn,
			wantFormat:  logging.JSONFormat,
			wantCall:    true,
		},
		{
			name:        "explicit level wins over environment",
			environment: map[string]string{"YCY_LOG_LEVEL": "warn", "YCY_LOG_FORMAT": "json"},
			arguments:   []string{"--log-level", "debug", "git", "heat"},
			wantLevel:   logging.Debug,
			wantFormat:  logging.JSONFormat,
			wantCall:    true,
		},
		{
			name:       "explicit format reaches runtime",
			arguments:  []string{"--log-format", "json", "git", "heat"},
			wantLevel:  logging.Info,
			wantFormat: logging.JSONFormat,
			wantCall:   true,
		},
		{
			name:       "verbose selects debug",
			arguments:  []string{"--verbose", "git", "heat"},
			wantLevel:  logging.Debug,
			wantFormat: logging.TextFormat,
			wantCall:   true,
		},
		{
			name:       "quiet selects error",
			arguments:  []string{"-q", "git", "heat"},
			wantLevel:  logging.Error,
			wantFormat: logging.TextFormat,
			wantCall:   true,
		},
		{
			name:       "conflicting controls stop before effect",
			arguments:  []string{"--log-level", "info", "--quiet", "git", "heat"},
			wantCode:   1,
			wantLevel:  logging.Info,
			wantFormat: logging.TextFormat,
			wantError:  "error: --log-level, --verbose, and --quiet are mutually exclusive\n",
		},
		{
			name:       "invalid format stops before effect",
			arguments:  []string{"--log-format", "yaml", "git", "heat"},
			wantCode:   1,
			wantLevel:  logging.Info,
			wantFormat: logging.TextFormat,
			wantError:  `error: invalid log format "yaml" (expected text or json)` + "\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			errors := &bytes.Buffer{}
			runtime := logging.NewRuntime(logging.Options{Writer: errors})
			called := false
			app, err := newTestApp(BuildInfo{Version: "0.0.0-dev"}, testDependencies{
				Out:         output,
				Err:         errors,
				Environment: func(key string) string { return test.environment[key] },
				Logging:     runtime,
				GitHeat: func(context.Context, githeat.Input) (githeat.Result, error) {
					called = true
					return githeat.Result{}, nil
				},
			})
			if err != nil {
				t.Fatalf("New returned an error: %v", err)
			}

			outcome := app.Execute(context.Background(), test.arguments)
			if outcome.Code != test.wantCode || errors.String() != test.wantError || called != test.wantCall || runtime.Level() != test.wantLevel || runtime.Format() != test.wantFormat {
				t.Fatalf("outcome = %#v, stderr = %q, called = %t, level = %v, format = %q", outcome, errors.String(), called, runtime.Level(), runtime.Format())
			}
		})
	}

	for _, arguments := range [][]string{
		{"--help", "--log-level", "invalid"},
		{"--version", "--log-level", "invalid"},
		{"completion", "bash", "--log-format", "yaml"},
	} {
		output := &bytes.Buffer{}
		errors := &bytes.Buffer{}
		app, err := newTestApp(BuildInfo{Version: "0.0.0-dev"}, testDependencies{
			Out:         output,
			Err:         errors,
			Environment: func(string) string { return "invalid" },
		})
		if err != nil {
			t.Fatalf("New returned an error: %v", err)
		}
		if outcome := app.Execute(context.Background(), arguments); outcome.Code != 0 || errors.Len() != 0 || output.Len() == 0 {
			t.Fatalf("discovery arguments %q outcome = %#v, stdout = %q, stderr = %q", arguments, outcome, output.String(), errors.String())
		}
	}
}

func TestRichTerminalDiscoveryPreservesVersionAndRawCompletion(t *testing.T) {
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	app, err := newTestApp(BuildInfo{Version: "0.0.0-dev"}, testDependencies{
		Out:     output,
		Err:     errors,
		Session: terminal.Session{Kind: terminal.RichInteractive},
		Logging: logging.NewRuntime(logging.Options{Writer: errors}),
	})
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}

	if outcome := app.Execute(context.Background(), []string{"--help"}); outcome.Code != 0 || !strings.Contains(output.String(), "Ycy command line interface") || errors.Len() != 0 {
		t.Fatalf("help outcome = %#v, stdout = %q, stderr = %q", outcome, output.String(), errors.String())
	}

	output.Reset()
	if outcome := app.Execute(context.Background(), []string{"--version"}); outcome.Code != 0 || output.String() != "0.0.0-dev\n" {
		t.Fatalf("version outcome = %#v, stdout = %q", outcome, output.String())
	}
	output.Reset()
	if outcome := app.Execute(context.Background(), []string{"completion", "bash"}); outcome.Code != 0 || !strings.Contains(output.String(), "__start_ycy") {
		t.Fatalf("completion outcome = %#v, stdout = %q", outcome, output.String())
	}
}

func TestParserRecoveryUsesOneActionableErrorLine(t *testing.T) {
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	called := false
	app, err := newTestApp(BuildInfo{Version: "0.0.0-dev"}, testDependencies{
		Out:     output,
		Err:     errors,
		Logging: logging.NewRuntime(logging.Options{Writer: errors}),
		GitHeat: func(context.Context, githeat.Input) (githeat.Result, error) {
			called = true
			return githeat.Result{}, nil
		},
	})
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}

	for _, testCase := range []struct {
		name      string
		arguments []string
		want      string
	}{
		{
			name:      "unique command candidate",
			arguments: []string{"exprot"},
			want:      "error: unknown command 'exprot'; did you mean 'export'? Run 'ycy export --help' for usage.\n",
		},
		{
			name:      "unique nested command candidate",
			arguments: []string{"config", "frok"},
			want:      "error: unknown command 'frok'; did you mean 'fork'? Run 'ycy config fork --help' for usage.\n",
		},
		{
			name:      "unique flag candidate",
			arguments: []string{"export", "env", "--merg"},
			want:      "error: unknown flag: --merg; did you mean '--merge'? Run 'ycy export env --help' for usage.\n",
		},
		{
			name:      "no candidate",
			arguments: []string{"unknown"},
			want:      "error: unknown command 'unknown'; Run 'ycy --help' for usage.\n",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			output.Reset()
			errors.Reset()
			called = false
			outcome := app.Execute(context.Background(), testCase.arguments)
			if outcome.Code != 1 || errors.String() != testCase.want || output.Len() != 0 || called {
				t.Fatalf("outcome = %#v, stdout = %q, stderr = %q, called = %t", outcome, output.String(), errors.String(), called)
			}
		})
	}
}

func TestParserRecoveryLeavesCommandErrorsUntouched(t *testing.T) {
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	app, err := newTestApp(BuildInfo{Version: "0.0.0-dev"}, testDependencies{
		Out:     output,
		Err:     errors,
		Logging: logging.NewRuntime(logging.Options{Writer: errors}),
		GitHeat: func(context.Context, githeat.Input) (githeat.Result, error) {
			return githeat.Result{}, stderrors.New("command validation failed")
		},
	})
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}

	outcome := app.Execute(context.Background(), []string{"git", "heat"})
	if outcome.Code != 1 || errors.String() != "error: command validation failed\n" || output.Len() != 0 {
		t.Fatalf("outcome = %#v, stdout = %q, stderr = %q", outcome, output.String(), errors.String())
	}
}

func TestConfigParentExposesMigratedForkGroup(t *testing.T) {
	app, output, errors, _ := testApp(t, nil)

	if outcome := app.Execute(context.Background(), []string{"config", "fork", "--help"}); outcome.Code != 0 || !strings.Contains(output.String(), "list") || !strings.Contains(output.String(), "add") || !strings.Contains(output.String(), "remove") {
		t.Fatalf("fork help outcome = %#v, stdout = %q", outcome, output.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"config", "fork", "list"}); outcome.Code != 0 || errors.Len() != 0 {
		t.Fatalf("config fork list outcome = %#v, stderr = %q", outcome, errors.String())
	}
}

func TestFSFoundationDoesNotExposeACommand(t *testing.T) {
	app, output, errors, _ := testApp(t, nil)

	if outcome := app.Execute(context.Background(), []string{"--help"}); outcome.Code != 0 || strings.Contains(output.String(), "\n  fs") {
		t.Fatalf("help outcome = %#v, stdout = %q", outcome, output.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"fs"}); outcome.Code != 1 || errors.String() != "error: unknown command 'fs'; did you mean 'rm'? Run 'ycy rm --help' for usage.\n" {
		t.Fatalf("fs outcome = %#v, stderr = %q", outcome, errors.String())
	}
}

func TestTunnelFoundationDoesNotExposeACommand(t *testing.T) {
	app, output, errors, _ := testApp(t, nil)

	if outcome := app.Execute(context.Background(), []string{"--help"}); outcome.Code != 0 || strings.Contains(output.String(), "\n  tunnel") {
		t.Fatalf("help outcome = %#v, stdout = %q", outcome, output.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"tunnel"}); outcome.Code != 1 || errors.String() != "error: unknown command 'tunnel'; Run 'ycy --help' for usage.\n" {
		t.Fatalf("tunnel outcome = %#v, stderr = %q", outcome, errors.String())
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
	app, err := newTestApp(BuildInfo{Version: "0.0.0-dev"}, testDependencies{
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
