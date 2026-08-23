package main

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestOSPulseGitRunnerUsesTheSharedArgvProcessBoundary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("host shell fixture is Unix-specific")
	}
	root := t.TempDir()
	argumentsPath := filepath.Join(root, "arguments")
	t.Setenv("PULSE_GIT_ARGUMENTS", argumentsPath)
	script := writeHeatGitScript(t, root, "git", "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$PULSE_GIT_ARGUMENTS\"\nprintf 'pulse stdout'\nprintf 'pulse stderr' >&2\n")
	runner := &osPulseGitRunner{executable: script}

	output, err := runner.Run(context.Background(), []string{"log", "--since=2026-08-23 00:00:00"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if output.ExitCode != 0 || string(output.Stdout) != "pulse stdout" || string(output.Stderr) != "pulse stderr" {
		t.Fatalf("output = %#v", output)
	}
	arguments, err := os.ReadFile(argumentsPath)
	if err != nil {
		t.Fatalf("read arguments: %v", err)
	}
	if got, want := string(arguments), "log\n--since=2026-08-23 00:00:00\n"; got != want {
		t.Fatalf("arguments = %q, want %q", got, want)
	}
}

func TestOSPulseGitRunnerPreservesMissingExecutableFailures(t *testing.T) {
	runner := &osPulseGitRunner{executable: filepath.Join(t.TempDir(), "missing-git")}
	_, err := runner.Run(context.Background(), nil)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Run() error = %v, want missing executable", err)
	}
}
