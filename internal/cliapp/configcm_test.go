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

func TestConfigCMListBindingPassesTypedInputAndExposesNoSibling(t *testing.T) {
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	runtime := logging.NewRuntime(logging.Options{Writer: errors})
	var inputs []configcm.Input
	app, err := New(BuildInfo{Version: "0.0.0-dev"}, Dependencies{
		Out:     output,
		Err:     errors,
		Logging: runtime,
		ConfigCMList: func(_ context.Context, input configcm.Input) (configcm.Result, error) {
			inputs = append(inputs, input)
			return configcm.Result{}, nil
		},
	})
	if err != nil {
		t.Fatalf("New() returned an error: %v", err)
	}

	outcome := app.Execute(context.Background(), []string{"--log-level", "warn", "config", "cm", "list"})
	if outcome.Code != 0 || outcome.Err != nil {
		t.Fatalf("list outcome = %#v, stderr = %q", outcome, errors.String())
	}
	if got, want := inputs, []configcm.Input{{}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("list inputs = %#v, want %#v", got, want)
	}
	if runtime.Level() != logging.Warn {
		t.Fatalf("log level = %v, want %v", runtime.Level(), logging.Warn)
	}

	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"config", "cm", "--help"}); outcome.Code != 0 || !strings.Contains(output.String(), "list") || strings.Contains(output.String(), "add") || strings.Contains(output.String(), "use") || strings.Contains(output.String(), "set") || strings.Contains(output.String(), "remove") || strings.Contains(output.String(), "test") {
		t.Fatalf("cm help outcome = %#v, stdout = %q", outcome, output.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"config", "cm", "add"}); outcome.Code != 1 || errors.String() != "error: unknown command 'add'; Run 'ycy config cm --help' for usage.\n" {
		t.Fatalf("absent sibling outcome = %#v, stderr = %q", outcome, errors.String())
	}
}

func TestConfigCMAddBindingPassesTypedRequestAndExposesOnlyRealLeaves(t *testing.T) {
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	runtime := logging.NewRuntime(logging.Options{Writer: errors})
	var addRequests []configcm.AddRequest
	app, err := New(BuildInfo{Version: "0.0.0-dev"}, Dependencies{
		Out:     output,
		Err:     errors,
		Logging: runtime,
		ConfigCMList: func(context.Context, configcm.Input) (configcm.Result, error) {
			return configcm.Result{}, nil
		},
		ConfigCMAdd: func(_ context.Context, request configcm.AddRequest) (configcm.AddResult, error) {
			addRequests = append(addRequests, request)
			return configcm.AddResult{}, nil
		},
	})
	if err != nil {
		t.Fatalf("New() returned an error: %v", err)
	}

	outcome := app.Execute(context.Background(), []string{"--log-level", "warn", "config", "cm", "add"})
	if outcome.Code != 0 || outcome.Err != nil {
		t.Fatalf("add outcome = %#v, stderr = %q", outcome, errors.String())
	}
	if got, want := addRequests, []configcm.AddRequest{{}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("add requests = %#v, want %#v", got, want)
	}
	if runtime.Level() != logging.Warn {
		t.Fatalf("log level = %v, want %v", runtime.Level(), logging.Warn)
	}

	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"config", "cm", "--help"}); outcome.Code != 0 || !strings.Contains(output.String(), "list") || !strings.Contains(output.String(), "add") || strings.Contains(output.String(), "use") || strings.Contains(output.String(), "set") || strings.Contains(output.String(), "remove") || strings.Contains(output.String(), "test") {
		t.Fatalf("cm help outcome = %#v, stdout = %q", outcome, output.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"config", "cm", "use"}); outcome.Code != 1 || errors.String() != "error: unknown command 'use'; Run 'ycy config cm --help' for usage.\n" {
		t.Fatalf("absent sibling outcome = %#v, stderr = %q", outcome, errors.String())
	}
}

func TestConfigCMUseBindingPassesTypedProfileAndExposesOnlyRealLeaves(t *testing.T) {
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	runtime := logging.NewRuntime(logging.Options{Writer: errors})
	var requests []configcm.UseRequest
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
		ConfigCMUse: func(_ context.Context, request configcm.UseRequest) (configcm.UseResult, error) {
			requests = append(requests, request)
			return configcm.UseResult{}, nil
		},
	})
	if err != nil {
		t.Fatalf("New() returned an error: %v", err)
	}

	outcome := app.Execute(context.Background(), []string{"--log-level", "warn", "config", "cm", "use", "work"})
	if outcome.Code != 0 || outcome.Err != nil {
		t.Fatalf("use outcome = %#v, stderr = %q", outcome, errors.String())
	}
	if got, want := requests, []configcm.UseRequest{{Profile: "work"}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("use requests = %#v, want %#v", got, want)
	}
	if runtime.Level() != logging.Warn {
		t.Fatalf("log level = %v, want %v", runtime.Level(), logging.Warn)
	}

	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"config", "cm", "use"}); outcome.Code != 1 || len(requests) != 1 || !strings.Contains(errors.String(), "accepts 1 arg(s), received 0") {
		t.Fatalf("missing profile outcome = %#v, requests = %#v, stderr = %q", outcome, requests, errors.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"config", "cm", "--help"}); outcome.Code != 0 || !strings.Contains(output.String(), "list") || !strings.Contains(output.String(), "add") || !strings.Contains(output.String(), "use") || strings.Contains(output.String(), "set") || strings.Contains(output.String(), "remove") || strings.Contains(output.String(), "test") {
		t.Fatalf("cm help outcome = %#v, stdout = %q", outcome, output.String())
	}
}

func TestConfigCMSetBindingPassesTypedRequestAndExposesOnlyRealLeaves(t *testing.T) {
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	runtime := logging.NewRuntime(logging.Options{Writer: errors})
	var requests []configcm.SetRequest
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
		ConfigCMSet: func(_ context.Context, request configcm.SetRequest) (configcm.SetResult, error) {
			requests = append(requests, request)
			return configcm.SetResult{}, nil
		},
	})
	if err != nil {
		t.Fatalf("New() returned an error: %v", err)
	}

	outcome := app.Execute(context.Background(), []string{"--log-level", "warn", "config", "cm", "set", "work", "timeoutMs", "1000suffix"})
	if outcome.Code != 0 || outcome.Err != nil {
		t.Fatalf("set outcome = %#v, stderr = %q", outcome, errors.String())
	}
	if got, want := requests, []configcm.SetRequest{{Profile: "work", Key: "timeoutMs", Value: "1000suffix"}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("set requests = %#v, want %#v", got, want)
	}
	if runtime.Level() != logging.Warn {
		t.Fatalf("log level = %v, want %v", runtime.Level(), logging.Warn)
	}

	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"config", "cm", "set", "work", "model"}); outcome.Code != 1 || len(requests) != 1 || !strings.Contains(errors.String(), "accepts 3 arg(s), received 2") {
		t.Fatalf("missing set operand outcome = %#v, requests = %#v, stderr = %q", outcome, requests, errors.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"config", "cm", "--help"}); outcome.Code != 0 || !strings.Contains(output.String(), "list") || !strings.Contains(output.String(), "add") || !strings.Contains(output.String(), "use") || !strings.Contains(output.String(), "set") || strings.Contains(output.String(), "remove") || strings.Contains(output.String(), "test") {
		t.Fatalf("cm help outcome = %#v, stdout = %q", outcome, output.String())
	}
}
