package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigCMUseStandaloneBinary(t *testing.T) {
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
	const apiKey = "api-key-that-must-not-escape"
	config := `{
  "salt": "bGVnYWN5LWNvbmZpZy1zYWx0",
  "fork": {"instances": {"github": {"host": "github.com", "type": "github", "token": "fork-token"}}},
  "cm": {"defaultProfile": "primary", "profiles": {
    "primary": {"baseURL": "https://primary.example/v1", "model": "primary-model", "apiKey": "` + apiKey + `"},
    "work": {"baseURL": "https://work.example/v1", "model": "work-model", "apiKey": "work-api-key"}
  }}
}`
	configPath := filepath.Join(configDirectory, "config.json")
	if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}
	environment := environmentWith(map[string]string{"HOME": home, "USERPROFILE": ""})

	output, err := runStandalone(binary, environment, "config", "cm", "use", "work")
	if err != nil || !strings.Contains(string(output), "Default CM profile set to work") || strings.Contains(string(output), apiKey) {
		t.Fatalf("config cm use = (%v, %q)", err, output)
	}
	contents, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	document := standaloneCMConfig(t, contents)
	if document.CM.DefaultProfile != "work" {
		t.Fatalf("default profile = %q, want work", document.CM.DefaultProfile)
	}
	if profile := standaloneCMProfile(t, contents, "primary"); profile.BaseURL != "https://primary.example/v1" || profile.Model != "primary-model" || profile.APIKey != apiKey {
		t.Fatalf("primary profile = %#v", profile)
	}
	if profile := standaloneCMProfile(t, contents, "work"); profile.BaseURL != "https://work.example/v1" || profile.Model != "work-model" || profile.APIKey != "work-api-key" {
		t.Fatalf("work profile = %#v", profile)
	}

	listOutput, err := runStandalone(binary, environment, "config", "cm", "list")
	if err != nil || !strings.Contains(string(listOutput), "* work") || strings.Contains(string(listOutput), apiKey) {
		t.Fatalf("config cm list after use = (%v, %q)", err, listOutput)
	}
	missingOutput, err := runStandalone(binary, environment, "config", "cm", "use", "missing")
	if err == nil || !strings.Contains(string(missingOutput), "error: CM profile not found: missing") || strings.Contains(string(missingOutput), "Default CM profile set") {
		t.Fatalf("missing profile = (%v, %q)", err, missingOutput)
	}
	helpOutput, err := runStandalone(binary, environment, "config", "cm", "--help")
	if err != nil || !strings.Contains(string(helpOutput), "list") || !strings.Contains(string(helpOutput), "add") || !strings.Contains(string(helpOutput), "use") || strings.Contains(string(helpOutput), "set") || strings.Contains(string(helpOutput), "remove") || strings.Contains(string(helpOutput), "test") {
		t.Fatalf("cm help = (%v, %q)", err, helpOutput)
	}
}
