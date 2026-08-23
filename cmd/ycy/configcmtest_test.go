package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	configcm "github.com/hackycy/hackycy-cli/internal/commands/config/cm"
)

func TestTerminalCMTestPresenterReportsOnlyTheSafeOutcomeFields(t *testing.T) {
	output := &bytes.Buffer{}
	presenter := terminalCMTestPresenter{output: output}
	presenter.Response("ok")
	presenter.Failure(configcm.TestDiagnostic{Provider: "work", BaseURL: "https://provider.test/v1", Model: "provider-model"})

	if got, want := output.String(), "Response: ok\nDone\nProvider: work\nBase URL: https://provider.test/v1\nModel: provider-model\n"; got != want {
		t.Fatalf("terminal CM test output = %q, want %q", got, want)
	}
}

func TestConfigCMTestStandaloneBinaryUsesOnlyTheLocalProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/chat/completions" || request.Header.Get("Authorization") != "Bearer test-api-key-that-must-not-escape" {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		var body struct {
			Model string `json:"model"`
		}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		if body.Model == "failure-model" {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusTooManyRequests)
			_, _ = fmt.Fprint(writer, `{"error":"test-api-key-that-must-not-escape"}`)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(writer, `{"choices":[{"message":{"content":"ok"}}]}`)
	}))
	defer server.Close()

	root := repositoryRoot(t)
	binary := filepath.Join(t.TempDir(), "ycy")
	build := exec.Command("go", "build", "-trimpath", "-o", binary, "./cmd/ycy")
	build.Dir = root
	build.Env = environmentWith(map[string]string{
		"CGO_ENABLED": "0",
		"GOTOOLCHAIN": "go1.26.7",
		"GOWORK":      "off",
	})
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build standalone binary: %v\n%s", err, output)
	}

	home := t.TempDir()
	configDirectory := filepath.Join(home, ".ycy-cli")
	if err := os.MkdirAll(configDirectory, 0o700); err != nil {
		t.Fatalf("create config directory: %v", err)
	}
	config := `{"cm":{"defaultProfile":"work","profiles":{"work":{"baseURL":"` + server.URL + `","model":"success-model"},"failure":{"baseURL":"` + server.URL + `","model":"failure-model"}}}}`
	if err := os.WriteFile(filepath.Join(configDirectory, "config.json"), []byte(config), 0o600); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}
	const apiKey = "test-api-key-that-must-not-escape"
	environment := environmentWith(map[string]string{
		"HOME":                     home,
		"USERPROFILE":              "",
		"YCY_CM_PROFILE":           "",
		"YCY_CM_BASE_URL":          "",
		"YCY_CM_MODEL":             "",
		"YCY_CM_API_KEY":           apiKey,
		"YCY_CM_TEMPERATURE":       "",
		"YCY_CM_TIMEOUT_MS":        "",
		"YCY_CM_MAX_OUTPUT_TOKENS": "",
	})

	successOutput, err := runStandalone(binary, environment, "config", "cm", "test")
	if err != nil || !strings.Contains(string(successOutput), "Response: ok") || !strings.Contains(string(successOutput), "Done") || strings.Contains(string(successOutput), apiKey) {
		t.Fatalf("config cm test success = (%v, %q)", err, successOutput)
	}

	failureOutput, err := runStandalone(binary, environment, "config", "cm", "test", "failure")
	if err == nil || !strings.Contains(string(failureOutput), "Provider: failure") || !strings.Contains(string(failureOutput), "Base URL: "+server.URL) || !strings.Contains(string(failureOutput), "Model: failure-model") || !strings.Contains(string(failureOutput), "error: 429 Too Many Requests: {\"error\":\"[REDACTED]\"}") || strings.Contains(string(failureOutput), apiKey) {
		t.Fatalf("config cm test failure = (%v, %q)", err, failureOutput)
	}

	helpOutput, err := runStandalone(binary, environment, "config", "cm", "--help")
	if err != nil || !strings.Contains(string(helpOutput), "list") || !strings.Contains(string(helpOutput), "add") || !strings.Contains(string(helpOutput), "use") || !strings.Contains(string(helpOutput), "set") || !strings.Contains(string(helpOutput), "remove") || !strings.Contains(string(helpOutput), "test") {
		t.Fatalf("cm help = (%v, %q)", err, helpOutput)
	}
}
