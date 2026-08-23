package cliapp

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	configcm "github.com/hackycy/hackycy-cli/internal/commands/config/cm"
	"github.com/hackycy/hackycy-cli/internal/logging"
)

func TestConfigCMTestBindingAcceptsAnOptionalTypedProfile(t *testing.T) {
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	runtime := logging.NewRuntime(logging.Options{Writer: errors})
	var requests []configcm.TestRequest
	app, err := New(BuildInfo{Version: "0.0.0-dev"}, Dependencies{
		Out:     output,
		Err:     errors,
		Logging: runtime,
		ConfigCMList: func(context.Context, configcm.Input) (configcm.Result, error) {
			return configcm.Result{}, nil
		},
		ConfigCMAdd: func(context.Context, configcm.AddRequest) (configcm.AddResult, error) {
			return configcm.AddResult{}, nil
		},
		ConfigCMUse: func(context.Context, configcm.UseRequest) (configcm.UseResult, error) {
			return configcm.UseResult{}, nil
		},
		ConfigCMSet: func(context.Context, configcm.SetRequest) (configcm.SetResult, error) {
			return configcm.SetResult{}, nil
		},
		ConfigCMRemove: func(context.Context, configcm.RemoveRequest) (configcm.RemoveResult, error) {
			return configcm.RemoveResult{}, nil
		},
		ConfigCMTest: func(_ context.Context, request configcm.TestRequest) (configcm.TestResult, error) {
			requests = append(requests, request)
			return configcm.TestResult{}, nil
		},
	})
	if err != nil {
		t.Fatalf("New() returned an error: %v", err)
	}

	for _, arguments := range [][]string{
		{"--log-level", "warn", "config", "cm", "test"},
		{"--log-level", "warn", "config", "cm", "test", "work"},
	} {
		if outcome := app.Execute(context.Background(), arguments); outcome.Code != 0 || outcome.Err != nil {
			t.Fatalf("test outcome = %#v, stderr = %q", outcome, errors.String())
		}
	}
	if got, want := requests, []configcm.TestRequest{{}, {Profile: "work"}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("test requests = %#v, want %#v", got, want)
	}
	if runtime.Level() != logging.Warn {
		t.Fatalf("log level = %v, want %v", runtime.Level(), logging.Warn)
	}

	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"config", "cm", "test", "work", "extra"}); outcome.Code != 1 || len(requests) != 2 || !strings.Contains(errors.String(), "accepts at most 1 arg(s), received 2") {
		t.Fatalf("extra profile outcome = %#v, requests = %#v, stderr = %q", outcome, requests, errors.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"config", "cm", "--help"}); outcome.Code != 0 || !strings.Contains(output.String(), "list") || !strings.Contains(output.String(), "add") || !strings.Contains(output.String(), "use") || !strings.Contains(output.String(), "set") || !strings.Contains(output.String(), "remove") || !strings.Contains(output.String(), "test") {
		t.Fatalf("cm help outcome = %#v, stdout = %q", outcome, output.String())
	}
}
