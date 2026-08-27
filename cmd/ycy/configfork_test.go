package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	configfork "github.com/hackycy/hackycy-cli/internal/commands/config/fork"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestTerminalForkListPresentationPreservesPlainAndAutomationResults(t *testing.T) {
	result := configfork.Result{Instances: []configfork.Instance{{
		Name:         "work",
		Host:         "gitlab.example",
		Scheme:       "https",
		Type:         "gitlab",
		TokenPreview: "MDEy***",
	}}}
	const want = "NAME  TYPE    SCHEME  HOST            TOKEN\nwork  gitlab  https   gitlab.example  MDEy***\n1 instance configured\n"

	for _, session := range []terminalexperience.Session{
		{Kind: terminalexperience.PlainInteractive},
		{Kind: terminalexperience.Automation},
	} {
		var output bytes.Buffer
		experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{Session: session, Output: &output})
		run := experience.Open(context.Background())
		if err := run.Present(terminalForkListDocument(session, result)); err != nil {
			t.Fatalf("Present() error = %v", err)
		}
		if err := run.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
		if got := output.String(); got != want {
			t.Fatalf("%v result = %q, want %q", session.Kind, got, want)
		}
		if terminaltest.ContainsTerminalControl(output.Bytes()) {
			t.Fatalf("%v result contains terminal control: %q", session.Kind, output.String())
		}
	}
}

func TestTerminalForkListPresentationUsesRichSemanticRoles(t *testing.T) {
	result := configfork.Result{Instances: []configfork.Instance{{
		Name:         "work",
		Host:         "gitlab.example",
		Scheme:       "https",
		Type:         "gitlab",
		TokenPreview: "MDEy***",
	}}}

	for _, testCase := range []struct {
		name    string
		session terminalexperience.Session
	}{
		{name: "color", session: terminalexperience.Session{Kind: terminalexperience.RichInteractive, Color: true}},
		{name: "no color", session: terminalexperience.Session{Kind: terminalexperience.RichInteractive}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			document := terminalForkListDocument(testCase.session, result)
			if got, want := []terminalexperience.VisualRole{
				document.Blocks[0].Role,
				document.Blocks[1].Role,
				document.Blocks[2].Role,
				document.Blocks[3].Role,
			}, []terminalexperience.VisualRole{
				terminalexperience.VisualRoleTitle,
				terminalexperience.VisualRoleMuted,
				terminalexperience.VisualRolePlain,
				terminalexperience.VisualRoleSuccess,
			}; !reflect.DeepEqual(got, want) {
				t.Fatalf("Rich roles = %#v, want %#v", got, want)
			}
			var output bytes.Buffer
			experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{Session: testCase.session, Output: &output})
			run := experience.Open(context.Background())
			if err := run.Present(document); err != nil {
				t.Fatalf("Present() error = %v", err)
			}
			if err := run.Close(); err != nil {
				t.Fatalf("Close() error = %v", err)
			}
			for _, expected := range []string{"Fork provider instances", "NAME  TYPE  SCHEME  HOST  TOKEN", "work", "gitlab.example", "MDEy***", "1 instance configured"} {
				if !strings.Contains(output.String(), expected) {
					t.Fatalf("Rich result = %q, missing %q", output.String(), expected)
				}
			}
			if terminaltest.ContainsTerminalControl(output.Bytes()) {
				t.Fatalf("non-terminal writer output contains terminal control: %q", output.String())
			}
		})
	}
}

func TestConfigForkListStandaloneBinary(t *testing.T) {
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
	if err != nil || !strings.Contains(string(helpOutput), "list") || !strings.Contains(string(helpOutput), "add") || !strings.Contains(string(helpOutput), "remove") {
		t.Fatalf("fork help = (%v, %q)", err, helpOutput)
	}
	missingOutput, err := runStandalone(binary, environment, "config", "fork", "remove")
	if err == nil || string(missingOutput) != "error: config fork remove requires an interactive terminal\n" {
		t.Fatalf("Automation removal = (%v, %q)", err, missingOutput)
	}
}

func TestConfigForkAddStandaloneBinary(t *testing.T) {
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
	environment := environmentWith(map[string]string{"HOME": home, "USERPROFILE": ""})
	output, err := runStandaloneWithInput(binary, environment, "work\nhttps://gitlab.example/path\n1\n2\nsecret-token\n", "config", "fork", "add")
	if err == nil || string(output) != "error: config fork add requires an interactive terminal\n" || strings.Contains(string(output), "secret-token") {
		t.Fatalf("Automation config fork add = (%v, %q)", err, output)
	}
	if _, err := os.Stat(filepath.Join(home, ".ycy-cli", "config.json")); !os.IsNotExist(err) {
		t.Fatalf("Automation config fork add wrote configuration: %v", err)
	}
}

func TestConfigForkRemoveStandaloneBinary(t *testing.T) {
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
	environment := environmentWith(map[string]string{"HOME": home, "USERPROFILE": ""})
	configPath := writeForkRemoveConfig(t, home)
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	output, err := runStandaloneWithInput(binary, environment, "1\ny\n", "config", "fork", "remove")
	if err == nil || string(output) != "error: config fork remove requires an interactive terminal\n" {
		t.Fatalf("Automation removal = (%v, %q)", err, output)
	}
	after, err := os.ReadFile(configPath)
	if err != nil || !bytes.Equal(after, before) {
		t.Fatalf("Automation removal changed config = (%v, %q)", err, after)
	}

	emptyHome := t.TempDir()
	emptyEnvironment := environmentWith(map[string]string{"HOME": emptyHome, "USERPROFILE": ""})
	emptyOutput, err := runStandalone(binary, emptyEnvironment, "config", "fork", "remove")
	if err != nil || string(emptyOutput) != "No instances configured\nNothing to remove\n" {
		t.Fatalf("empty removal = (%v, %q)", err, emptyOutput)
	}
	if _, err := os.Stat(filepath.Join(emptyHome, ".ycy-cli", "config.json")); !os.IsNotExist(err) {
		t.Fatalf("empty removal created config: %v", err)
	}

	helpOutput, err := runStandalone(binary, environment, "config", "fork", "--help")
	if err != nil || !strings.Contains(string(helpOutput), "list") || !strings.Contains(string(helpOutput), "add") || !strings.Contains(string(helpOutput), "remove") {
		t.Fatalf("fork help = (%v, %q)", err, helpOutput)
	}
}

func runStandalone(binary string, environment []string, arguments ...string) ([]byte, error) {
	command := exec.Command(resolveStandaloneBinary(binary), arguments...)
	command.Env = environment
	return command.CombinedOutput()
}

func runStandaloneWithInput(binary string, environment []string, input string, arguments ...string) ([]byte, error) {
	command := exec.Command(resolveStandaloneBinary(binary), arguments...)
	command.Env = environment
	command.Stdin = strings.NewReader(input)
	return command.CombinedOutput()
}

type standaloneForkDocument struct {
	Fork struct {
		Instances map[string]standaloneForkInstanceRecord `json:"instances"`
	} `json:"fork"`
}

type standaloneForkInstanceRecord struct {
	Host   string `json:"host"`
	Scheme string `json:"scheme"`
	Type   string `json:"type"`
	Token  string `json:"token"`
}

func standaloneForkInstance(t *testing.T, contents []byte, name string) standaloneForkInstanceRecord {
	t.Helper()
	var document standaloneForkDocument
	if err := json.Unmarshal(contents, &document); err != nil {
		t.Fatalf("decode config: %v", err)
	}
	instance, found := document.Fork.Instances[name]
	if !found {
		t.Fatalf("config omitted %q: %q", name, contents)
	}
	return instance
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
