package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigForkListStandaloneBinary(t *testing.T) {
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
	const ciphertext = "MDEyMzQ1Njc4OWFiY2RlZg==:tag:ciphertext"
	config := `{
  "salt": "bGVnYWN5LWNvbmZpZy1zYWx0",
  "fork": {"instances": {
    "work": {"host": "gitlab.example", "type": "gitlab", "token": "MDEyMzQ1Njc4OWFiY2RlZg==:tag:ciphertext"},
    "personal": {"host": "github.example", "scheme": "http", "type": "github", "token": "QWVy:second:ciphertext"}
  }}
}`
	if err := os.WriteFile(filepath.Join(configDirectory, "config.json"), []byte(config), 0o600); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}
	environment := environmentWith(map[string]string{"HOME": home, "USERPROFILE": ""})

	output, err := runStandalone(binary, environment, "config", "fork", "list")
	if err != nil {
		t.Fatalf("config fork list: %v\n%s", err, output)
	}
	for _, field := range []string{"work", "gitlab", "https", "gitlab.example", "MDEy***", "personal", "github", "http", "github.example", "QWVy***"} {
		if !strings.Contains(string(output), field) {
			t.Fatalf("list output missing %q: %q", field, output)
		}
	}
	if strings.Index(string(output), "work") > strings.Index(string(output), "personal") {
		t.Fatalf("list output did not preserve stored order: %q", output)
	}
	if strings.Contains(string(output), ciphertext) {
		t.Fatalf("list output exposed the full ciphertext: %q", output)
	}

	helpOutput, err := runStandalone(binary, environment, "config", "fork", "--help")
	if err != nil || !strings.Contains(string(helpOutput), "list") || strings.Contains(string(helpOutput), "add") || strings.Contains(string(helpOutput), "remove") {
		t.Fatalf("fork help = (%v, %q)", err, helpOutput)
	}
	missingOutput, err := runStandalone(binary, environment, "config", "fork", "add")
	if err == nil || string(missingOutput) != "error: unknown command 'add'\n" {
		t.Fatalf("absent sibling = (%v, %q)", err, missingOutput)
	}
}

func runStandalone(binary string, environment []string, arguments ...string) ([]byte, error) {
	command := exec.Command(binary, arguments...)
	command.Env = environment
	return command.CombinedOutput()
}

func environmentWith(overrides map[string]string) []string {
	environment := make([]string, 0, len(os.Environ())+len(overrides))
	for _, entry := range os.Environ() {
		key, _, found := strings.Cut(entry, "=")
		if !found {
			continue
		}
		if _, replaced := overrides[key]; !replaced {
			environment = append(environment, entry)
		}
	}
	for key, value := range overrides {
		environment = append(environment, key+"="+value)
	}
	return environment
}
