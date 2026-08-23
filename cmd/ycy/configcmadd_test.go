package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	configcm "github.com/hackycy/hackycy-cli/internal/commands/config/cm"
)

func TestTerminalCMAddPrompterValidatesTextWithoutTrimmingAcceptedInput(t *testing.T) {
	output := &bytes.Buffer{}
	prompter := newTerminalCMAddPrompter(strings.NewReader("\n https://provider.example/v1/// \n"), output)
	question := configcm.AddTextPrompt{
		Message:     "OpenAI-compatible base URL",
		Placeholder: "https://api.openai.com/v1",
		Validate: func(value string) error {
			if strings.TrimSpace(value) == "" {
				return errors.New("Base URL is required")
			}
			return nil
		},
	}

	value, cancelled := prompter.Text(question)

	if cancelled || value != " https://provider.example/v1/// " {
		t.Fatalf("Text() = (%q, %t)", value, cancelled)
	}
	if !strings.Contains(output.String(), "OpenAI-compatible base URL") || !strings.Contains(output.String(), "https://api.openai.com/v1") || !strings.Contains(output.String(), "Base URL is required") {
		t.Fatalf("output = %q", output.String())
	}
}

func TestTerminalCMAddPrompterReadsPipedPasswordWithoutWritingIt(t *testing.T) {
	output := &bytes.Buffer{}
	prompter := newTerminalCMAddPrompter(strings.NewReader("secret-api-key\n"), output)

	value, cancelled := prompter.Password(configcm.AddTextPrompt{Message: "API key", Validate: func(string) error { return nil }})

	if cancelled || value != "secret-api-key" {
		t.Fatalf("Password() = (%q, %t)", value, cancelled)
	}
	if strings.Contains(output.String(), "secret-api-key") {
		t.Fatalf("password prompt exposed input: %q", output.String())
	}
}

func TestTerminalCMAddPrompterTreatsEOFAsCancellation(t *testing.T) {
	prompter := newTerminalCMAddPrompter(strings.NewReader(""), &bytes.Buffer{})
	value, cancelled := prompter.Text(configcm.AddTextPrompt{Message: "Model", Validate: func(string) error { return nil }})
	if !cancelled || value != "" {
		t.Fatalf("Text() = (%q, %t)", value, cancelled)
	}
}

func TestTerminalCMAddPresenterWritesOutcomeMessages(t *testing.T) {
	output := &bytes.Buffer{}
	presenter := terminalCMAddPresenter{output: output}

	presenter.Cancel("Cancelled")
	presenter.Success("Profile work added")

	if got, want := output.String(), "Cancelled\nProfile work added\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestConfigCMAddStandaloneBinary(t *testing.T) {
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
	environment := environmentWith(map[string]string{"HOME": home, "USERPROFILE": ""})
	const apiKey = "api-key-that-must-not-escape"
	output, err := runStandaloneWithInput(binary, environment, "work\n https://provider.example/v1/// \n gpt-4.1-mini \n"+apiKey+"\n", "config", "cm", "add")
	if err != nil {
		t.Fatalf("config cm add: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "Profile work added") || strings.Contains(string(output), apiKey) {
		t.Fatalf("add output = %q", output)
	}

	configPath := filepath.Join(home, ".ycy-cli", "config.json")
	contents, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if strings.Contains(string(contents), apiKey) {
		t.Fatalf("config contains plaintext API key: %q", contents)
	}
	profile := standaloneCMProfile(t, contents, "work")
	if profile.BaseURL != "https://provider.example/v1" || profile.Model != "gpt-4.1-mini" || profile.APIKey == "" || profile.APIKey == apiKey || strings.Count(profile.APIKey, ":") != 2 {
		t.Fatalf("persisted profile = %#v", profile)
	}
	document := standaloneCMConfig(t, contents)
	if document.CM.DefaultProfile != "work" {
		t.Fatalf("default profile = %q, want work", document.CM.DefaultProfile)
	}

	listOutput, err := runStandalone(binary, environment, "config", "cm", "list")
	if err != nil {
		t.Fatalf("config cm list after add: %v\n%s", err, listOutput)
	}
	for _, field := range []string{"* work", "gpt-4.1-mini", "https://provider.example/v1"} {
		if !strings.Contains(string(listOutput), field) {
			t.Fatalf("list output missing %q: %q", field, listOutput)
		}
	}
	if strings.Contains(string(listOutput), apiKey) || strings.Contains(string(listOutput), profile.APIKey) {
		t.Fatalf("list output exposed API key material: %q", listOutput)
	}

	overwriteOutput, err := runStandaloneWithInput(binary, environment, "work\nhttps://replacement.example/v1///\nreplacement-model\nreplacement-api-key\n", "config", "cm", "add")
	if err != nil || !strings.Contains(string(overwriteOutput), "Profile work added") || strings.Contains(string(overwriteOutput), "replacement-api-key") {
		t.Fatalf("overwrite output = (%v, %q)", err, overwriteOutput)
	}
	contents, err = os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read overwritten config: %v", err)
	}
	profile = standaloneCMProfile(t, contents, "work")
	if profile.BaseURL != "https://replacement.example/v1" || profile.Model != "replacement-model" || profile.APIKey == "" || profile.APIKey == "replacement-api-key" || strings.Contains(string(contents), "replacement-api-key") {
		t.Fatalf("overwritten profile = %#v", profile)
	}

	cancelledHome := t.TempDir()
	cancelledEnvironment := environmentWith(map[string]string{"HOME": cancelledHome, "USERPROFILE": ""})
	cancelledOutput, err := runStandaloneWithInput(binary, cancelledEnvironment, "", "config", "cm", "add")
	if err != nil || !strings.Contains(string(cancelledOutput), "Cancelled") {
		t.Fatalf("cancelled add = (%v, %q)", err, cancelledOutput)
	}
	if _, err := os.Stat(filepath.Join(cancelledHome, ".ycy-cli", "config.json")); !os.IsNotExist(err) {
		t.Fatalf("cancelled add created configuration: %v", err)
	}

	failureHome := t.TempDir()
	if err := os.WriteFile(filepath.Join(failureHome, ".ycy-cli"), []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("create failing config root: %v", err)
	}
	failureEnvironment := environmentWith(map[string]string{"HOME": failureHome, "USERPROFILE": ""})
	failureOutput, err := runStandaloneWithInput(binary, failureEnvironment, "work\nhttps://provider.example\nmodel\n"+apiKey+"\n", "config", "cm", "add")
	if err == nil || !strings.Contains(string(failureOutput), "error:") || strings.Contains(string(failureOutput), "Profile work added") || strings.Contains(string(failureOutput), apiKey) {
		t.Fatalf("save failure = (%v, %q)", err, failureOutput)
	}
}

type standaloneCMConfigDocument struct {
	CM struct {
		DefaultProfile string                             `json:"defaultProfile"`
		Profiles       map[string]standaloneCMProfileData `json:"profiles"`
	} `json:"cm"`
}

type standaloneCMProfileData struct {
	BaseURL string `json:"baseURL"`
	Model   string `json:"model"`
	APIKey  string `json:"apiKey"`
}

func standaloneCMConfig(t *testing.T, contents []byte) standaloneCMConfigDocument {
	t.Helper()
	var document standaloneCMConfigDocument
	if err := json.Unmarshal(contents, &document); err != nil {
		t.Fatalf("decode config: %v", err)
	}
	return document
}

func standaloneCMProfile(t *testing.T, contents []byte, name string) standaloneCMProfileData {
	t.Helper()
	document := standaloneCMConfig(t, contents)
	profile, found := document.CM.Profiles[name]
	if !found {
		t.Fatalf("config omitted %q: %q", name, contents)
	}
	return profile
}
