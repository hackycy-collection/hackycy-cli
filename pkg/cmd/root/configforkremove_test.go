package root

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	configfork "github.com/hackycy/hackycy-cli/internal/commands/config/fork"
	"github.com/hackycy/hackycy-cli/internal/logging"
)

func TestConfigForkRemoveBindingPassesTypedRequestAndExposesOnlyTheRealLeaf(t *testing.T) {
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	runtime := logging.NewRuntime(logging.Options{Writer: errors})
	var requests []configfork.RemoveRequest
	app, err := newTestApp(BuildInfo{Version: "0.0.0-dev"}, testDependencies{
		Out:     output,
		Err:     errors,
		Logging: runtime,
		ConfigForkRemove: func(_ context.Context, request configfork.RemoveRequest) (configfork.RemoveResult, error) {
			requests = append(requests, request)
			return configfork.RemoveResult{}, nil
		},
	})
	if err != nil {
		t.Fatalf("New() returned an error: %v", err)
	}

	outcome := app.Execute(context.Background(), []string{"--log-level", "warn", "config", "fork", "remove"})
	if outcome.Code != 0 || outcome.Err != nil {
		t.Fatalf("remove outcome = %#v, stderr = %q", outcome, errors.String())
	}
	if got, want := requests, []configfork.RemoveRequest{{}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("remove requests = %#v, want %#v", got, want)
	}
	if runtime.Level() != logging.Warn {
		t.Fatalf("log level = %v, want %v", runtime.Level(), logging.Warn)
	}

	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"config", "fork", "--help"}); outcome.Code != 0 || !strings.Contains(output.String(), "remove") || strings.Contains(output.String(), "list") || strings.Contains(output.String(), "add") {
		t.Fatalf("fork help outcome = %#v, stdout = %q", outcome, output.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"config", "fork", "list"}); outcome.Code != 1 || errors.String() != "error: unknown command 'list'; Run 'ycy config fork --help' for usage.\n" {
		t.Fatalf("absent sibling outcome = %#v, stderr = %q", outcome, errors.String())
	}
}
