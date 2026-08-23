package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	runcommand "github.com/hackycy/hackycy-cli/internal/commands/run"
)

func TestTerminalRunPrompterAndPresenterUseTheRunContract(t *testing.T) {
	output := &bytes.Buffer{}
	prompter := newTerminalRunPrompter(strings.NewReader("2\n1\n"), output)
	script, cancelled := prompter.SelectScript(runcommand.ScriptPrompt{
		Message: "Select a script to run:",
		Options: []runcommand.ScriptChoice{
			{Value: "check", Label: "check", Hint: "go test ./..."},
			{Value: "build", Label: "build", Hint: "go build ./cmd/ycy"},
		},
	})
	if cancelled || script != "build" {
		t.Fatalf("SelectScript() = (%q, %t)", script, cancelled)
	}
	manager, cancelled := prompter.SelectPackageManager(runcommand.PackageManagerPrompt{
		Message: "Select a package manager:",
		Options: []runcommand.PackageManagerChoice{{Value: runcommand.PackageManagerExternal, Label: string(runcommand.PackageManagerExternal)}},
	})
	if cancelled || manager != runcommand.PackageManagerExternal {
		t.Fatalf("SelectPackageManager() = (%q, %t)", manager, cancelled)
	}
	if !strings.Contains(output.String(), "build - go build ./cmd/ycy") || !strings.Contains(output.String(), "Select a package manager:") {
		t.Fatalf("prompt output = %q", output.String())
	}

	cancellation := newTerminalRunPrompter(strings.NewReader("cancel\n"), &bytes.Buffer{})
	_, cancelled = cancellation.SelectScript(runcommand.ScriptPrompt{Options: []runcommand.ScriptChoice{{Value: "check"}}})
	if !cancelled {
		t.Fatal("SelectScript() did not treat cancel as cancellation")
	}

	presented := &bytes.Buffer{}
	presenter := terminalRunPresenter{output: presented}
	presenter.Intro("Run Script")
	presenter.Info(string(runcommand.PackageManagerExternal) + " run check")
	presenter.Blank()
	presenter.Cancel("Operation cancelled.")
	want := "HACKYCY CLI\n\nRun Script\n" + string(runcommand.PackageManagerExternal) + " run check\n\nOperation cancelled.\n"
	if presented.String() != want {
		t.Fatalf("presentation = %q, want %q", presented.String(), want)
	}
}

func TestRunStandaloneBinaryPreservesProjectExecutionAndParserBehavior(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("host shell fixture is Unix-specific")
	}
	repository := repositoryRoot(t)
	binary := filepath.Join(t.TempDir(), "ycy")
	build := exec.Command("go", "build", "-trimpath", "-o", binary, "./cmd/ycy")
	build.Dir = repository
	build.Env = environmentWith(map[string]string{
		"CGO_ENABLED": "0",
		"GOTOOLCHAIN": "go1.26.7",
		"GOWORK":      "off",
	})
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build standalone binary: %v\n%s", err, output)
	}

	root := t.TempDir()
	project := filepath.Join(root, "project")
	writeStandaloneRunFile(t, project, "package.json", `{"scripts":{"check":"echo check"}}`)
	writeStandaloneRunFile(t, project, "b"+"un"+".lock", "")
	binDirectory := filepath.Join(root, "bin")
	argumentsPath := filepath.Join(root, "arguments")
	workingDirectoryPath := filepath.Join(root, "working-directory")
	manager := filepath.Join(binDirectory, string(runcommand.PackageManagerExternal))
	managerScript := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$RUN_ARGUMENTS\"\npwd > \"$RUN_WORKING_DIRECTORY\"\nprintf 'external child output'\nif [ -n \"$RUN_EXIT\" ]; then\n  exit \"$RUN_EXIT\"\nfi\n"
	if err := os.MkdirAll(binDirectory, 0o700); err != nil {
		t.Fatalf("create manager directory: %v", err)
	}
	if err := os.WriteFile(manager, []byte(managerScript), 0o700); err != nil {
		t.Fatalf("write manager fixture: %v", err)
	}

	environment := environmentWith(map[string]string{
		"HOME":                  t.TempDir(),
		"USERPROFILE":           "",
		"PATH":                  binDirectory + string(os.PathListSeparator) + os.Getenv("PATH"),
		"RUN_ARGUMENTS":         argumentsPath,
		"RUN_WORKING_DIRECTORY": workingDirectoryPath,
	})
	output, err := runRunStandalone(binary, project, environment, "1\n1\n", "run")
	if err != nil || !strings.Contains(string(output), "Run Script") || !strings.Contains(string(output), "external child output") {
		t.Fatalf("successful run = (%v, %q)", err, output)
	}
	assertRunProcessFile(t, argumentsPath, "run\ncheck\n")
	resolvedProject, err := filepath.EvalSymlinks(project)
	if err != nil {
		t.Fatalf("resolve project path: %v", err)
	}
	assertRunProcessFile(t, workingDirectoryPath, resolvedProject+"\n")

	for _, arguments := range [][]string{{"run", ".", "--flag"}, {"run", "--flag", "value"}, {"run", "arg1", "arg2"}, {"run", "--", "arg1", "arg2"}} {
		output, err = runRunStandalone(binary, project, environment, "", arguments...)
		if exitCode(err) != 1 || !strings.Contains(string(output), "accepts at most 1 arg(s)") {
			t.Fatalf("arguments %q = (%v, %q)", arguments, err, output)
		}
	}

	output, err = runRunStandalone(binary, project, environment, "", "run", "--help")
	if err != nil || !strings.Contains(string(output), "Run package.json scripts") {
		t.Fatalf("run help = (%v, %q)", err, output)
	}

	output, err = runRunStandalone(binary, project, environment, "1\n1\n", "run", ".", "--log-level", "warn")
	if err != nil || !strings.Contains(string(output), "external child output") {
		t.Fatalf("leaf log level = (%v, %q)", err, output)
	}

	output, err = runRunStandalone(binary, project, environment, "", "run", "--flag")
	if exitCode(err) != 1 || !strings.Contains(string(output), "No package.json found in current directory.") {
		t.Fatalf("option-like path = (%v, %q)", err, output)
	}

	missingEnvironment := environmentWith(map[string]string{
		"HOME":                  t.TempDir(),
		"USERPROFILE":           "",
		"PATH":                  filepath.Join(root, "missing-bin"),
		"RUN_ARGUMENTS":         argumentsPath,
		"RUN_WORKING_DIRECTORY": workingDirectoryPath,
	})
	output, err = runRunStandalone(binary, project, missingEnvironment, "1\n1\n", "run")
	if exitCode(err) != 1 || !strings.Contains(string(output), "executable file not found") {
		t.Fatalf("missing executable = (%v, %q)", err, output)
	}

	exitEnvironment := environmentWith(map[string]string{
		"HOME":                  t.TempDir(),
		"USERPROFILE":           "",
		"PATH":                  binDirectory + string(os.PathListSeparator) + os.Getenv("PATH"),
		"RUN_ARGUMENTS":         argumentsPath,
		"RUN_WORKING_DIRECTORY": workingDirectoryPath,
		"RUN_EXIT":              "7",
	})
	output, err = runRunStandalone(binary, project, exitEnvironment, "1\n1\n", "run")
	if exitCode(err) != 7 || !strings.Contains(string(output), "external child output") {
		t.Fatalf("child exit = (%v, %q)", err, output)
	}
}

func runRunStandalone(binary, directory string, environment []string, input string, arguments ...string) ([]byte, error) {
	command := exec.Command(binary, arguments...)
	command.Dir = directory
	command.Env = environment
	command.Stdin = strings.NewReader(input)
	return command.CombinedOutput()
}

func writeStandaloneRunFile(t *testing.T, directory, name, contents string) {
	t.Helper()
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatalf("create package directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, name), []byte(contents), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var exited *exec.ExitError
	if errors.As(err, &exited) {
		return exited.ExitCode()
	}
	return -1
}
