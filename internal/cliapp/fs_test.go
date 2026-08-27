package cliapp

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	fscommand "github.com/hackycy/hackycy-cli/internal/commands/fs"
)

func TestFSBindingPassesTypedInputAndEnvironmentPrecedence(t *testing.T) {
	app, output, errors, inputs := fsTestApp(t, map[string]string{
		"YCY_FS_SESSION_DIR":           "environment-sessions",
		"YCY_FS_SESSION_IDLE_DAYS":     "3",
		"YCY_FS_CHUNKED_UPLOAD":        "true",
		"YCY_FS_UPLOAD_CHUNK_SIZE_MIB": "4",
	})
	configured := app.Execute(context.Background(), []string{
		"fs", "--port", "00000", "--address", "127.0.0.1", "--manage", "--safe-html",
		"--account", "Alice:password", "--account=Bob:password", "--session-dir", "selected-sessions",
		"--session-idle-days", "2", "--chunked-upload=false", "--upload-chunk-size", "16", "chosen-directory",
	})
	if configured.Code != 0 || configured.Err != nil {
		t.Fatalf("configured outcome = %#v, stderr = %q", configured, errors.String())
	}
	environment := app.Execute(context.Background(), []string{"fs", "environment-directory"})
	if environment.Code != 0 || environment.Err != nil {
		t.Fatalf("environment outcome = %#v, stderr = %q", environment, errors.String())
	}
	want := []fscommand.Input{
		{
			Directory:          "chosen-directory",
			Address:            "127.0.0.1",
			ManagementEnabled:  true,
			SafeHTML:           true,
			Accounts:           []string{"Alice:password", "Bob:password"},
			SessionDirectory:   "selected-sessions",
			SessionIdleTimeout: 2 * 24 * time.Hour,
			UploadChunkSize:    16 * 1024 * 1024,
		},
		{
			Directory:          "environment-directory",
			Port:               1204,
			Address:            "0.0.0.0",
			SessionDirectory:   "environment-sessions",
			SessionIdleTimeout: 3 * 24 * time.Hour,
			ChunkedUploads:     true,
			UploadChunkSize:    4 * 1024 * 1024,
		},
	}
	if !reflect.DeepEqual(*inputs, want) {
		t.Fatalf("inputs = %#v, want %#v", *inputs, want)
	}
	if output.Len() != 0 || errors.Len() != 0 {
		t.Fatalf("binding output = %q stderr = %q", output.String(), errors.String())
	}
}

func TestFSBindingRejectsInvalidConfigurationBeforeInvokingTheHandler(t *testing.T) {
	for _, testCase := range []struct {
		arguments []string
		want      string
	}{
		{arguments: []string{"fs", "--port=-1"}, want: "'-1' is not a valid port"},
		{arguments: []string{"fs", "--port=65536"}, want: "Port must be between 0 and 65535"},
		{arguments: []string{"fs", "--session-idle-days=0"}, want: "'0' is not a valid positive session idle day count"},
		{arguments: []string{"fs", "--session-idle-days=999999999999"}, want: "'999999999999' is not a valid positive session idle day count"},
		{arguments: []string{"fs", "--upload-chunk-size=3"}, want: "'3' is not a valid upload chunk size; use 4-16 MiB"},
		{arguments: []string{"fs", "first", "second"}, want: "accepts at most 1 arg(s)"},
	} {
		t.Run(strings.Join(testCase.arguments, " "), func(t *testing.T) {
			app, _, errors, inputs := fsTestApp(t, nil)
			outcome := app.Execute(context.Background(), testCase.arguments)
			if outcome.Code != 1 || len(*inputs) != 0 || !strings.Contains(errors.String(), testCase.want) {
				t.Fatalf("outcome = %#v inputs = %#v stderr = %q", outcome, *inputs, errors.String())
			}
		})
	}

	app, _, errors, inputs := fsTestApp(t, map[string]string{"YCY_FS_CHUNKED_UPLOAD": "yes"})
	outcome := app.Execute(context.Background(), []string{"fs"})
	if outcome.Code != 1 || len(*inputs) != 0 || !strings.Contains(errors.String(), "'yes' is not a valid chunked-upload value") {
		t.Fatalf("environment outcome = %#v inputs = %#v stderr = %q", outcome, *inputs, errors.String())
	}
}

func TestFSBindingExposesOnlyAnInjectedHandlerAndItsHelp(t *testing.T) {
	app, output, errors, _ := testApp(t, nil)
	if outcome := app.Execute(context.Background(), []string{"--help"}); outcome.Code != 0 || strings.Contains(output.String(), "fs [directory]") {
		t.Fatalf("unregistered root help outcome = %#v, stdout = %q, stderr = %q", outcome, output.String(), errors.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"fs"}); outcome.Code != 1 || errors.String() != "error: unknown command 'fs'; Run 'ycy --help' for usage.\n" {
		t.Fatalf("unregistered fs outcome = %#v, stderr = %q", outcome, errors.String())
	}

	app, output, errors, _ = fsTestApp(t, nil)
	if outcome := app.Execute(context.Background(), []string{"fs", "--help"}); outcome.Code != 0 || !strings.Contains(output.String(), "Browse a directory in a browser") || !strings.Contains(output.String(), "--account") || !strings.Contains(output.String(), "--upload-chunk-size") || strings.Contains(output.String(), "--public") {
		t.Fatalf("registered fs help outcome = %#v, stdout = %q, stderr = %q", outcome, output.String(), errors.String())
	}
}

func fsTestApp(t *testing.T, environment map[string]string) (*App, *bytes.Buffer, *bytes.Buffer, *[]fscommand.Input) {
	t.Helper()
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	inputs := []fscommand.Input{}
	app, err := New(BuildInfo{Version: "0.0.0-dev"}, Dependencies{
		Out: output,
		Err: errors,
		Environment: func(key string) string {
			return environment[key]
		},
		FS: func(_ context.Context, input fscommand.Input) (fscommand.Result, error) {
			inputs = append(inputs, input)
			return fscommand.Result{}, nil
		},
	})
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}
	return app, output, errors, &inputs
}
