package cliapp

import (
	"bytes"
	"context"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/hackycy/hackycy-cli/internal/commands/tunnel"
	"github.com/hackycy/hackycy-cli/internal/logging"
)

func TestTunnelServerBindingResolvesCLIEnvironmentAndGlobalLogging(t *testing.T) {
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	runtime := logging.NewRuntime(logging.Options{Writer: errors})
	configuredDataDirectory := filepath.Join(t.TempDir(), "configured")
	environmentDataDirectory := filepath.Join(t.TempDir(), "environment")
	environment := map[string]string{
		"YCY_TUNNEL_ADDRESS":            "0.0.0.0",
		"YCY_TUNNEL_CONTROL_PORT":       "7500",
		"YCY_TUNNEL_FRP_PORT":           "7000",
		"YCY_TUNNEL_HTTP_PORT":          "8080",
		"YCY_TUNNEL_PORT_RANGE":         "20000-20100",
		"YCY_TUNNEL_ADVERTISE_FRP_ADDR": "environment.example.test:7555",
		"YCY_TUNNEL_DATA_DIR":           environmentDataDirectory,
		"YCY_TUNNEL_SESSION_IDLE_DAYS":  "7",
		"YCY_TUNNEL_ADMIN_USER":         "ops-admin",
		"YCY_TUNNEL_ADMIN_PASSWORD":     "environment-password",
		"YCY_TUNNEL_FRP_TOKEN":          "environment-token",
	}
	var inputs []tunnel.ServerConfig
	app, err := New(BuildInfo{Version: "0.0.0-dev"}, Dependencies{
		Out:               output,
		Err:               errors,
		Logging:           runtime,
		Environment:       tunnelEnvironmentValue(environment),
		EnvironmentLookup: tunnelEnvironmentLookup(environment),
		TunnelServer: func(_ context.Context, input tunnel.ServerConfig) error {
			inputs = append(inputs, input)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	configured := app.Execute(context.Background(), []string{
		"--log-level", "warn", "tunnel", "server",
		"--address", "127.0.0.1", "--control-port", "7501", "--frp-port", "7001", "--http-port", "8081",
		"--port-range", "21000-21100", "--advertise-frp-addr", "[2001:db8::1]:7443",
		"--data-dir", configuredDataDirectory, "--session-idle-days", "8",
	})
	if configured.Code != 0 || configured.Err != nil {
		t.Fatalf("configured tunnel server outcome = %#v, stderr = %q", configured, errors.String())
	}
	if runtime.Level() != logging.Warn {
		t.Fatalf("configured global log level = %v, want %v", runtime.Level(), logging.Warn)
	}
	environmentOutcome := app.Execute(context.Background(), []string{"tunnel", "server"})
	if environmentOutcome.Code != 0 || environmentOutcome.Err != nil {
		t.Fatalf("environment tunnel server outcome = %#v, stderr = %q", environmentOutcome, errors.String())
	}
	want := []tunnel.ServerConfig{
		{
			Settings: tunnel.ServerHTTPServerSettings{
				Address:          "127.0.0.1",
				ControlPort:      7501,
				FRPPort:          7001,
				HTTPPort:         8081,
				PortRange:        tunnel.ServerHTTPPortRange{Start: 21000, End: 21100},
				AdvertiseFRPAddr: &tunnel.ServerHTTPFRPAddress{Host: "2001:db8::1", Port: 7443},
				DataDir:          configuredDataDirectory,
				AdminUser:        "ops-admin",
			},
			AdminPassword:       "environment-password",
			SessionIdleLifetime: 8 * 24 * time.Hour,
			FRPToken:            "environment-token",
		},
		{
			Settings: tunnel.ServerHTTPServerSettings{
				Address:          "0.0.0.0",
				ControlPort:      7500,
				FRPPort:          7000,
				HTTPPort:         8080,
				PortRange:        tunnel.ServerHTTPPortRange{Start: 20000, End: 20100},
				AdvertiseFRPAddr: &tunnel.ServerHTTPFRPAddress{Host: "environment.example.test", Port: 7555},
				DataDir:          environmentDataDirectory,
				AdminUser:        "ops-admin",
			},
			AdminPassword:       "environment-password",
			SessionIdleLifetime: 7 * 24 * time.Hour,
			FRPToken:            "environment-token",
		},
	}
	if !reflect.DeepEqual(inputs, want) {
		t.Fatalf("Tunnel server inputs = %#v, want %#v", inputs, want)
	}
	if output.Len() != 0 || errors.Len() != 0 {
		t.Fatalf("Tunnel server binding output = %q stderr = %q", output.String(), errors.String())
	}
}

func TestTunnelServerBindingRejectsInvalidConfigurationBeforeInvokingTheHandler(t *testing.T) {
	for _, testCase := range []struct {
		name        string
		environment map[string]string
		arguments   []string
		message     string
	}{
		{name: "missing password", arguments: []string{"tunnel", "server"}, message: "YCY_TUNNEL_ADMIN_PASSWORD"},
		{name: "invalid port", environment: map[string]string{"YCY_TUNNEL_ADMIN_PASSWORD": "environment-password"}, arguments: []string{"tunnel", "server", "--control-port=0"}, message: "Control port must be an integer"},
		{name: "empty configured token", environment: map[string]string{"YCY_TUNNEL_ADMIN_PASSWORD": "environment-password", "YCY_TUNNEL_FRP_TOKEN": ""}, arguments: []string{"tunnel", "server"}, message: "FRP Token must not be empty"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			app, output, errors, calls := tunnelServerTestApp(t, testCase.environment)
			outcome := app.Execute(context.Background(), testCase.arguments)
			if outcome.Code != 1 || *calls != 0 || !strings.Contains(errors.String(), testCase.message) || output.Len() != 0 {
				t.Fatalf("outcome = %#v calls = %d stdout = %q stderr = %q", outcome, *calls, output.String(), errors.String())
			}
		})
	}
}

func TestTunnelConnectBindingPreservesExplicitFlagsAndConfiguresGlobalLogging(t *testing.T) {
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	runtime := logging.NewRuntime(logging.Options{Writer: errors})
	var inputs []tunnel.ClientOptionInput
	app, err := New(BuildInfo{Version: "0.0.0-dev"}, Dependencies{
		Out:     output,
		Err:     errors,
		Logging: runtime,
		TunnelConnect: func(_ context.Context, input tunnel.ClientOptionInput) error {
			inputs = append(inputs, input)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	for index, arguments := range [][]string{
		{"--log-level", "warn", "tunnel", "connect", "--server", "http://control.example.test", "--token", "cli-token"},
		{"tunnel", "connect", "--server="},
		{"tunnel", "connect"},
	} {
		if outcome := app.Execute(context.Background(), arguments); outcome.Code != 0 || outcome.Err != nil {
			t.Fatalf("arguments %q outcome = %#v, stderr = %q", arguments, outcome, errors.String())
		}
		if index == 0 && runtime.Level() != logging.Warn {
			t.Fatalf("global log level = %v, want %v", runtime.Level(), logging.Warn)
		}
	}
	if len(inputs) != 3 {
		t.Fatalf("connect handler calls = %d, want 3", len(inputs))
	}
	if got, want := requiredTunnelConnectOption(inputs[0].Server), "http://control.example.test"; got != want {
		t.Errorf("first server = %q, want %q", got, want)
	}
	if got, want := requiredTunnelConnectOption(inputs[0].Token), "cli-token"; got != want {
		t.Errorf("first token = %q, want %q", got, want)
	}
	if inputs[1].Server == nil || *inputs[1].Server != "" || inputs[1].Token != nil {
		t.Errorf("explicit empty server input = %#v", inputs[1])
	}
	if inputs[2].Server != nil || inputs[2].Token != nil {
		t.Errorf("absent connect options = %#v", inputs[2])
	}
	if output.Len() != 0 || errors.Len() != 0 {
		t.Fatalf("connect binding output = %q stderr = %q", output.String(), errors.String())
	}
}

func TestTunnelConnectBindingRegistersOnlyItsOwnLeafAndFlags(t *testing.T) {
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	app, err := New(BuildInfo{Version: "0.0.0-dev"}, Dependencies{
		Out: output,
		Err: errors,
		TunnelConnect: func(context.Context, tunnel.ClientOptionInput) error {
			return nil
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if outcome := app.Execute(context.Background(), []string{"tunnel", "--help"}); outcome.Code != 0 || !strings.Contains(output.String(), "connect") || strings.Contains(output.String(), "server") {
		t.Fatalf("tunnel help outcome = %#v, stdout = %q", outcome, output.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"tunnel", "connect", "--help"}); outcome.Code != 0 || !strings.Contains(output.String(), "--server") || !strings.Contains(output.String(), "--token") || strings.Contains(output.String(), "--control-port") || !strings.Contains(output.String(), "Global Flags:\n      --log-level") {
		t.Fatalf("connect help outcome = %#v, stdout = %q", outcome, output.String())
	}
}

func TestTunnelServerBindingExposesOnlyTheIntegratedServerLeaf(t *testing.T) {
	app, output, errors, _ := testApp(t, nil)
	if outcome := app.Execute(context.Background(), []string{"--help"}); outcome.Code != 0 || strings.Contains(output.String(), "\n  tunnel") {
		t.Fatalf("unregistered root help outcome = %#v, stdout = %q", outcome, output.String())
	}

	app, output, errors, _ = tunnelServerTestApp(t, map[string]string{"YCY_TUNNEL_ADMIN_PASSWORD": "environment-password"})
	if outcome := app.Execute(context.Background(), []string{"--help"}); outcome.Code != 0 || !strings.Contains(output.String(), "tunnel") {
		t.Fatalf("registered root help outcome = %#v, stdout = %q", outcome, output.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"tunnel", "--help"}); outcome.Code != 0 || !strings.Contains(output.String(), "server") || strings.Contains(output.String(), "connect") {
		t.Fatalf("tunnel help outcome = %#v, stdout = %q", outcome, output.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"tunnel", "server", "--help"}); outcome.Code != 0 || !strings.Contains(output.String(), "--control-port") || !strings.Contains(output.String(), "--session-idle-days") || !strings.Contains(output.String(), "Global Flags:\n      --log-level") || strings.Contains(strings.Split(output.String(), "Global Flags:")[0], "--log-level") {
		t.Fatalf("tunnel server help outcome = %#v, stdout = %q", outcome, output.String())
	}
	output.Reset()
	errors.Reset()
	if outcome := app.Execute(context.Background(), []string{"tunnel", "connect"}); outcome.Code != 1 || errors.String() != "error: unknown command 'connect'\n" {
		t.Fatalf("unintegrated tunnel connect outcome = %#v, stderr = %q", outcome, errors.String())
	}
}

func tunnelServerTestApp(t *testing.T, environment map[string]string) (*App, *bytes.Buffer, *bytes.Buffer, *int) {
	t.Helper()
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	calls := 0
	app, err := New(BuildInfo{Version: "0.0.0-dev"}, Dependencies{
		Out:               output,
		Err:               errors,
		Environment:       tunnelEnvironmentValue(environment),
		EnvironmentLookup: tunnelEnvironmentLookup(environment),
		TunnelServer: func(context.Context, tunnel.ServerConfig) error {
			calls++
			return nil
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return app, output, errors, &calls
}

func requiredTunnelConnectOption(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func tunnelEnvironmentValue(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func tunnelEnvironmentLookup(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
