package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigCMSetStandaloneBinary(t *testing.T) {
	root := repositoryRoot(t)
	binary := standaloneBinaryOutputPath(filepath.Join(t.TempDir(), "ycy"))
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
	const originalAPIKey = "initial-api-key-that-must-not-escape"
	const rotatedAPIKey = "rotated-api-key-that-must-not-escape"
	config := `{
  "salt": "bGVnYWN5LWNvbmZpZy1zYWx0",
  "fork": {"instances": {"github": {"host": "github.com", "type": "github", "token": "MDEyMzQ1Njc4OWFiY2RlZg==:dc4+2YzE3mJaY01o7OyLtw==:QvEGhYbu7lpp4lc2/N0a8IXumQ=="}}},
  "cm": {"defaultProfile": "work", "profiles": {
    "work": {"baseURL": "https://old.example/v1", "model": "old-model", "apiKey": "` + originalAPIKey + `", "temperature": 0.2, "timeoutMs": 300000, "maxOutputTokens": 1000}
  }}
}`
	configPath := filepath.Join(configDirectory, "config.json")
	if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}
	environment := environmentWith(map[string]string{"HOME": home, "USERPROFILE": ""})

	for _, update := range []struct {
		key   string
		value string
	}{
		{key: "baseURL", value: " https://provider.example/v2/// "},
		{key: "model", value: " next-model "},
		{key: "apiKey", value: rotatedAPIKey},
		{key: "temperature", value: "0b1"},
		{key: "timeoutMs", value: "1000suffix"},
		{key: "maxOutputTokens", value: "32.9"},
	} {
		output, err := runStandalone(binary, environment, "config", "cm", "set", "work", update.key, update.value)
		if err != nil || !strings.Contains(string(output), "Profile work updated") || strings.Contains(string(output), update.value) {
			t.Fatalf("config cm set %s = (%v, %q)", update.key, err, output)
		}
	}

	contents, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	profile := standaloneCMProfile(t, contents, "work")
	if profile.BaseURL != "https://provider.example/v2" || profile.Model != "next-model" || profile.APIKey == "" || profile.APIKey == rotatedAPIKey || strings.Count(profile.APIKey, ":") != 2 || profile.Temperature != 1 || profile.TimeoutMS != 1000 || profile.MaxOutputTokens != 32 {
		t.Fatalf("persisted profile = %#v", profile)
	}
	if strings.Contains(string(contents), originalAPIKey) || strings.Contains(string(contents), rotatedAPIKey) {
		t.Fatalf("config contains plaintext API key: %q", contents)
	}
	if document := standaloneCMConfig(t, contents); document.CM.DefaultProfile != "work" {
		t.Fatalf("default profile = %q, want work", document.CM.DefaultProfile)
	}

	listOutput, err := runStandalone(binary, environment, "config", "cm", "list")
	if err != nil || !strings.Contains(string(listOutput), "* work") || !strings.Contains(string(listOutput), "next-model") || strings.Contains(string(listOutput), rotatedAPIKey) || strings.Contains(string(listOutput), profile.APIKey) {
		t.Fatalf("config cm list after set = (%v, %q)", err, listOutput)
	}
	for _, arguments := range [][]string{
		{"config", "cm", "set", "missing", "model", "next"},
		{"config", "cm", "set", "work", "unsupported", "value"},
		{"config", "cm", "set", "work", "temperature", "2.1"},
	} {
		output, err := runStandalone(binary, environment, arguments...)
		if err == nil || !strings.Contains(string(output), "error:") || strings.Contains(string(output), "Profile work updated") || strings.Contains(string(output), rotatedAPIKey) {
			t.Fatalf("%q = (%v, %q)", arguments, err, output)
		}
	}

	helpOutput, err := runStandalone(binary, environment, "config", "cm", "--help")
	if err != nil || !strings.Contains(string(helpOutput), "list") || !strings.Contains(string(helpOutput), "add") || !strings.Contains(string(helpOutput), "use") || !strings.Contains(string(helpOutput), "set") || !strings.Contains(string(helpOutput), "remove") || !strings.Contains(string(helpOutput), "test") {
		t.Fatalf("cm help = (%v, %q)", err, helpOutput)
	}
}
