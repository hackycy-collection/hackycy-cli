package root

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	configcm "github.com/hackycy/hackycy-cli/internal/commands/config/cm"
	"github.com/hackycy/hackycy-cli/internal/logging"
)

func TestConfigCMRemoveBindingPassesTypedProfileAndExposesOnlyRealLeaves(t *testing.T) {
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	runtime := logging.NewRuntime(logging.Options{Writer: errors})
	var requests []configcm.RemoveRequest
	app, err := newTestApp(BuildInfo{Version: "0.0.0-dev"}, testDependencies{
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
		ConfigCMRemove: func(_ context.Context, request configcm.RemoveRequest) (configcm.RemoveResult, error) {
			requests = append(requests, request)
			return configcm.RemoveResult{}, nil
		},
	})
	if err != nil {
		t.Fatalf("New() returned an error: %v", err)
	}

	outcome := app.Execute(context.Background(), []string{"--log-level", "warn", "config", "cm", "remove", "work"})
	if outcome.Code != 0 || outcome.Err != nil {
		t.Fatalf("remove outcome = %#v, stderr = %q", outcome, errors.String())
	}
	if got, want := requests, []configcm.RemoveRequest{{Profile: "work"}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("remove requests = %#v, want %#v", got, want)
	}
	if runtime.Level() != logging.Warn {
		t.Fatalf("log level = %v, want %v", runtime.Level(), logging.Warn)
	}

	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"config", "cm", "remove"}); outcome.Code != 1 || len(requests) != 1 || !strings.Contains(errors.String(), "accepts 1 arg(s), received 0") {
		t.Fatalf("missing profile outcome = %#v, requests = %#v, stderr = %q", outcome, requests, errors.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"config", "cm", "--help"}); outcome.Code != 0 || !strings.Contains(output.String(), "list") || !strings.Contains(output.String(), "add") || !strings.Contains(output.String(), "use") || !strings.Contains(output.String(), "set") || !strings.Contains(output.String(), "remove") || strings.Contains(output.String(), "test") {
		t.Fatalf("cm help outcome = %#v, stdout = %q", outcome, output.String())
	}
}
