package main

import (
	"encoding/json"
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
	if err != nil || !strings.Contains(string(helpOutput), "list") || !strings.Contains(string(helpOutput), "add") || strings.Contains(string(helpOutput), "remove") {
		t.Fatalf("fork help = (%v, %q)", err, helpOutput)
	}
	missingOutput, err := runStandalone(binary, environment, "config", "fork", "remove")
	if err == nil || string(missingOutput) != "error: unknown command 'remove'\n" {
		t.Fatalf("absent sibling = (%v, %q)", err, missingOutput)
	}
}

func TestConfigForkAddStandaloneBinary(t *testing.T) {
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
	output, err := runStandaloneWithInput(binary, environment, "work\nhttps://gitlab.example/path\n1\n2\nsecret-token\n", "config", "fork", "add")
	if err != nil {
		t.Fatalf("config fork add: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "Instance work (https://gitlab.example/path) added successfully") || strings.Contains(string(output), "secret-token") {
		t.Fatalf("add output = %q", output)
	}

	configPath := filepath.Join(home, ".ycy-cli", "config.json")
	contents, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if strings.Contains(string(contents), "secret-token") {
		t.Fatalf("config contains plaintext token: %q", contents)
	}
	instance := standaloneForkInstance(t, contents, "work")
	if instance.Host != "https://gitlab.example/path" || instance.Scheme != "http" || instance.Type != "gitlab" || instance.Token == "" || instance.Token == "secret-token" || strings.Count(instance.Token, ":") != 2 {
		t.Fatalf("persisted instance = %#v", instance)
	}

	listOutput, err := runStandalone(binary, environment, "config", "fork", "list")
	if err != nil {
		t.Fatalf("config fork list after add: %v\n%s", err, listOutput)
	}
	for _, field := range []string{"work", "gitlab", "https", "gitlab.example"} {
		if !strings.Contains(string(listOutput), field) {
			t.Fatalf("normalized list output missing %q: %q", field, listOutput)
		}
	}
	if strings.Contains(string(listOutput), "secret-token") || strings.Contains(string(listOutput), instance.Token) {
		t.Fatalf("list output exposed token material: %q", listOutput)
	}

	overwriteOutput, err := runStandaloneWithInput(binary, environment, "work\nreplacement.example\n2\n2\nreplacement-token\n", "config", "fork", "add")
	if err != nil {
		t.Fatalf("config fork add overwrite: %v\n%s", err, overwriteOutput)
	}
	if !strings.Contains(string(overwriteOutput), "Instance work (replacement.example) added successfully") || strings.Contains(string(overwriteOutput), "replacement-token") {
		t.Fatalf("overwrite output = %q", overwriteOutput)
	}
	contents, err = os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read overwritten config: %v", err)
	}
	instance = standaloneForkInstance(t, contents, "work")
	if instance.Host != "replacement.example" || instance.Scheme != "http" || instance.Type != "github" || instance.Token == "" || instance.Token == "replacement-token" || strings.Contains(string(contents), "replacement-token") {
		t.Fatalf("overwritten instance = %#v", instance)
	}

	cancelledHome := t.TempDir()
	cancelledEnvironment := environmentWith(map[string]string{"HOME": cancelledHome, "USERPROFILE": ""})
	cancelledOutput, err := runStandaloneWithInput(binary, cancelledEnvironment, "", "config", "fork", "add")
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
	failureOutput, err := runStandaloneWithInput(binary, failureEnvironment, "work\ngitlab.example\n1\n1\nsecret-token\n", "config", "fork", "add")
	if err == nil || !strings.Contains(string(failureOutput), "error:") || strings.Contains(string(failureOutput), "added successfully") || strings.Contains(string(failureOutput), "secret-token") {
		t.Fatalf("save failure = (%v, %q)", err, failureOutput)
	}
}

func runStandalone(binary string, environment []string, arguments ...string) ([]byte, error) {
	command := exec.Command(binary, arguments...)
	command.Env = environment
	return command.CombinedOutput()
}

func runStandaloneWithInput(binary string, environment []string, input string, arguments ...string) ([]byte, error) {
	command := exec.Command(binary, arguments...)
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
