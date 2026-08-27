package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	cmcommand "github.com/hackycy/hackycy-cli/internal/commands/git/cm"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
)

func TestOSCMGitRunnerPassesArgvAndInputAndCapturesOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("host shell fixture is Unix-specific")
	}
	root := t.TempDir()
	argumentsPath := filepath.Join(root, "arguments")
	inputPath := filepath.Join(root, "input")
	t.Setenv("CM_GIT_ARGUMENTS", argumentsPath)
	t.Setenv("CM_GIT_INPUT", inputPath)
	script := writeHeatGitScript(t, root, "git", "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$CM_GIT_ARGUMENTS\"\ncat > \"$CM_GIT_INPUT\"\nprintf 'child stdout'\nprintf 'child stderr' >&2\n")

	runner := &osCMGitRunner{executable: script}
	output, err := runner.RunInput(context.Background(), []string{"-C", "/repo", "cat-file", "--batch"}, []byte("HEAD:package.json\n"))
	if err != nil {
		t.Fatalf("RunInput() error = %v", err)
	}
	if output.ExitCode != 0 || string(output.Stdout) != "child stdout" || string(output.Stderr) != "child stderr" {
		t.Fatalf("RunInput() output = %#v", output)
	}
	arguments, err := os.ReadFile(argumentsPath)
	if err != nil {
		t.Fatalf("read arguments: %v", err)
	}
	if got, want := string(arguments), "-C\n/repo\ncat-file\n--batch\n"; got != want {
		t.Fatalf("arguments = %q, want %q", got, want)
	}
	input, err := os.ReadFile(inputPath)
	if err != nil {
		t.Fatalf("read input: %v", err)
	}
	if got, want := string(input), "HEAD:package.json\n"; got != want {
		t.Fatalf("input = %q, want %q", got, want)
	}
}

func TestOSCMGitRunnerMapsNonzeroExitAndCancellation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("host shell fixture is Unix-specific")
	}
	runner := &osCMGitRunner{executable: writeHeatGitScript(t, t.TempDir(), "git", "#!/bin/sh\nprintf 'fatal output' >&2\nexit 9\n")}
	output, err := runner.Run(context.Background(), nil)
	if err != nil || output.ExitCode != 9 || string(output.Stderr) != "fatal output" {
		t.Fatalf("Run() = (%#v, %v)", output, err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = runner.RunInput(cancelled, nil, []byte("ignored"))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RunInput() error = %v, want context cancellation", err)
	}
}

func TestOSCMGitRunnerSupportsGitCatFileBatch(t *testing.T) {
	repository := newGitCMRepository(t)
	writeGitCMFile(t, filepath.Join(repository, "package.json"), "{\"name\":\"after\"}\n")
	runGitCM(t, repository, "add", "package.json")

	output, err := newOSCMGitRunner().RunInput(context.Background(), []string{"-C", repository, "cat-file", "--batch"}, []byte("HEAD:package.json\n:package.json\n"))
	if err != nil || output.ExitCode != 0 {
		t.Fatalf("RunInput() = (%#v, %v)", output, err)
	}
	if !strings.Contains(string(output.Stdout), "{\"name\":\"before\"}") || !strings.Contains(string(output.Stdout), "{\"name\":\"after\"}") {
		t.Fatalf("cat-file output = %q", output.Stdout)
	}
}

func TestOSCMSnapshotFileSystemUsesOSFileOperations(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file.txt")
	writeGitCMFile(t, path, "contents")
	fileSystem := osCMSnapshotFileSystem{}
	info, err := fileSystem.Lstat(path)
	if err != nil || info.Size() != int64(len("contents")) {
		t.Fatalf("Lstat() = (%v, %v)", info, err)
	}
	reader, err := fileSystem.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer reader.Close()
	opened, err := io.ReadAll(reader)
	if err != nil || string(opened) != "contents" {
		t.Fatalf("Open() contents = %q, %v", opened, err)
	}
	contents, err := fileSystem.ReadFile(path)
	if err != nil || string(contents) != "contents" {
		t.Fatalf("ReadFile() = %q, %v", contents, err)
	}
}

func TestGitCMHandlerComposesTheProductionAdaptersForNoChanges(t *testing.T) {
	repository := newGitCMRepository(t)
	withGitCMWorkingDirectory(t, repository)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", "")
	output := &bytes.Buffer{}
	diagnostics := &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Session:     terminalexperience.Session{Kind: terminalexperience.PlainInteractive},
		Input:       strings.NewReader(""),
		Output:      output,
		Diagnostics: diagnostics,
	})

	result, err := newGitCMHandler(experience)(context.Background(), cmcommand.Input{})
	if err != nil || !result.NoChanges || result.NoChangeScope != cmcommand.ScopeAllUncommitted {
		t.Fatalf("handler() = (%#v, %v)", result, err)
	}
	if got, want := output.String(), "No uncommitted changes.\n"; got != want {
		t.Fatalf("handler output = %q, want %q", got, want)
	}
}

func newGitCMRepository(t *testing.T) string {
	t.Helper()
	repository := t.TempDir()
	runGitCM(t, repository, "init")
	runGitCM(t, repository, "config", "user.email", "fixture@example.test")
	runGitCM(t, repository, "config", "user.name", "Fixture")
	writeGitCMFile(t, filepath.Join(repository, "package.json"), "{\"name\":\"before\"}\n")
	runGitCM(t, repository, "add", "package.json")
	runGitCM(t, repository, "commit", "-m", "chore: initialize fixture")
	return repository
}

func runGitCM(t *testing.T, directory string, arguments ...string) {
	t.Helper()
	command := exec.Command("git", arguments...)
	command.Dir = directory
	command.Env = environmentWith(map[string]string{"GIT_CONFIG_NOSYSTEM": "1"})
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(arguments, " "), err, output)
	}
}

func writeGitCMFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func withGitCMWorkingDirectory(t *testing.T, directory string) {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir(directory); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
}
