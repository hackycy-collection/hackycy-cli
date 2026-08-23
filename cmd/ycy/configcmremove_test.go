package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	configcm "github.com/hackycy/hackycy-cli/internal/commands/config/cm"
)

func TestTerminalCMRemovePrompterConfirmsAndDefaultsToNo(t *testing.T) {
	output := &bytes.Buffer{}
	prompter := newTerminalCMRemovePrompter(strings.NewReader("yes\n"), output)

	confirmed, cancelled := prompter.Confirm(configcm.RemoveConfirmPrompt{Message: "Remove CM profile \"work\"?"})

	if !confirmed || cancelled {
		t.Fatalf("Confirm() = (%t, %t), want confirmation", confirmed, cancelled)
	}
	if !strings.Contains(output.String(), "Remove CM profile \"work\"? [y/N]:") {
		t.Fatalf("confirmation output = %q", output.String())
	}

	defaultPrompter := newTerminalCMRemovePrompter(strings.NewReader("\n"), &bytes.Buffer{})
	confirmed, cancelled = defaultPrompter.Confirm(configcm.RemoveConfirmPrompt{Message: "Remove CM profile \"work\"?"})
	if confirmed || cancelled {
		t.Fatalf("default Confirm() = (%t, %t), want negative confirmation", confirmed, cancelled)
	}
}

func TestTerminalCMRemovePrompterRetriesInvalidInputAndTreatsEOFAsCancellation(t *testing.T) {
	output := &bytes.Buffer{}
	prompter := newTerminalCMRemovePrompter(strings.NewReader("invalid\ny\n"), output)

	confirmed, cancelled := prompter.Confirm(configcm.RemoveConfirmPrompt{Message: "Remove CM profile \"work\"?"})

	if !confirmed || cancelled || !strings.Contains(output.String(), "Invalid confirmation") {
		t.Fatalf("Confirm() = (%t, %t), output = %q", confirmed, cancelled, output.String())
	}

	eofPrompter := newTerminalCMRemovePrompter(strings.NewReader(""), &bytes.Buffer{})
	confirmed, cancelled = eofPrompter.Confirm(configcm.RemoveConfirmPrompt{Message: "Remove CM profile \"work\"?"})
	if confirmed || !cancelled {
		t.Fatalf("EOF Confirm() = (%t, %t), want cancellation", confirmed, cancelled)
	}
}

func TestConfigCMRemoveStandaloneBinary(t *testing.T) {
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
	const primaryKey = "primary-api-key-that-must-not-escape"
	const workKey = "work-api-key-that-must-not-escape"
	const lastKey = "last-api-key-that-must-not-escape"
	config := strings.Join([]string{
		"{",
		"  \"salt\": \"bGVnYWN5LWNvbmZpZy1zYWx0\",",
		"  \"fork\": {\"instances\": {\"github\": {\"host\": \"github.com\", \"type\": \"github\", \"token\": \"fork-token\"}}},",
		"  \"cm\": {\"defaultProfile\": \"work\", \"profiles\": {",
		"    \"primary\": {\"baseURL\": \"https://primary.example/v1\", \"model\": \"primary-model\", \"apiKey\": \"" + primaryKey + "\"},",
		"    \"work\": {\"baseURL\": \"https://work.example/v1\", \"model\": \"work-model\", \"apiKey\": \"" + workKey + "\"},",
		"    \"last\": {\"baseURL\": \"https://last.example/v1\", \"model\": \"last-model\", \"apiKey\": \"" + lastKey + "\"}",
		"  }}",
		"}",
	}, "\n")
	configPath := filepath.Join(configDirectory, "config.json")
	if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}
	environment := environmentWith(map[string]string{"HOME": home, "USERPROFILE": ""})

	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config before cancellation: %v", err)
	}
	for _, input := range []string{"", "\n"} {
		output, err := runStandaloneWithInput(binary, environment, input, "config", "cm", "remove", "work")
		if err != nil || !strings.Contains(string(output), "Cancelled") {
			t.Fatalf("non-mutating removal input %q = (%v, %q)", input, err, output)
		}
		after, readErr := os.ReadFile(configPath)
		if readErr != nil || string(after) != string(before) {
			t.Fatalf("non-mutating removal changed config = (%v, %q)", readErr, after)
		}
	}

	output, err := runStandaloneWithInput(binary, environment, "yes\n", "config", "cm", "remove", "work")
	if err != nil || !strings.Contains(string(output), "Profile work removed") || strings.Contains(string(output), workKey) {
		t.Fatalf("default removal = (%v, %q)", err, output)
	}
	contents, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config after default removal: %v", err)
	}
	document := standaloneCMConfig(t, contents)
	if document.CM.DefaultProfile != "primary" || len(document.CM.Profiles) != 2 {
		t.Fatalf("CM after default removal = %#v", document.CM)
	}
	if _, found := document.CM.Profiles["work"]; found || document.CM.Profiles["primary"].APIKey != primaryKey || document.CM.Profiles["last"].APIKey != lastKey || !strings.Contains(string(contents), "github.com") {
		t.Fatalf("default removal did not preserve unrelated state: %s", contents)
	}

	output, err = runStandaloneWithInput(binary, environment, "y\n", "config", "cm", "remove", "last")
	if err != nil || !strings.Contains(string(output), "Profile last removed") {
		t.Fatalf("nondefault removal = (%v, %q)", err, output)
	}
	contents, err = os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config after nondefault removal: %v", err)
	}
	document = standaloneCMConfig(t, contents)
	if document.CM.DefaultProfile != "primary" || len(document.CM.Profiles) != 1 {
		t.Fatalf("CM after nondefault removal = %#v", document.CM)
	}

	output, err = runStandaloneWithInput(binary, environment, "yes\n", "config", "cm", "remove", "primary")
	if err != nil || !strings.Contains(string(output), "Profile primary removed") {
		t.Fatalf("last removal = (%v, %q)", err, output)
	}
	contents, err = os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config after last removal: %v", err)
	}
	document = standaloneCMConfig(t, contents)
	if document.CM.DefaultProfile != "" || len(document.CM.Profiles) != 0 {
		t.Fatalf("CM after last removal = %#v", document.CM)
	}

	missingOutput, err := runStandaloneWithInput(binary, environment, "yes\n", "config", "cm", "remove", "missing")
	if err == nil || !strings.Contains(string(missingOutput), "error: CM profile not found: missing") || strings.Contains(string(missingOutput), "Profile missing removed") {
		t.Fatalf("missing profile = (%v, %q)", err, missingOutput)
	}
	helpOutput, err := runStandalone(binary, environment, "config", "cm", "--help")
	if err != nil || !strings.Contains(string(helpOutput), "list") || !strings.Contains(string(helpOutput), "add") || !strings.Contains(string(helpOutput), "use") || !strings.Contains(string(helpOutput), "set") || !strings.Contains(string(helpOutput), "remove") || strings.Contains(string(helpOutput), "test") {
		t.Fatalf("cm help = (%v, %q)", err, helpOutput)
	}
}
