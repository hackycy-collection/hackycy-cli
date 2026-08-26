package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	rmcommand "github.com/hackycy/hackycy-cli/internal/commands/rm"
)

func TestTerminalRMPrompterUsesLegacyDefaultsAndSelections(t *testing.T) {
	confirmationOutput := &bytes.Buffer{}
	confirmation := newTerminalRMPrompter(strings.NewReader("\n"), confirmationOutput)
	confirmed, cancelled := confirmation.ConfirmExplicit(rmcommand.ExplicitConfirmationPrompt{Message: "Delete 1 item?"})
	if confirmed || cancelled || !strings.Contains(confirmationOutput.String(), "Delete 1 item? [y/N]:") {
		t.Fatalf("default confirmation = (%t, %t, %q)", confirmed, cancelled, confirmationOutput.String())
	}

	eofConfirmation := newTerminalRMPrompter(strings.NewReader(""), &bytes.Buffer{})
	confirmed, cancelled = eofConfirmation.ConfirmExplicit(rmcommand.ExplicitConfirmationPrompt{Message: "Delete 1 item?"})
	if confirmed || !cancelled {
		t.Fatalf("EOF confirmation = (%t, %t)", confirmed, cancelled)
	}

	actions := []rmcommand.SmartAction{{ID: "one", Label: "One"}, {ID: "two", Label: "Two"}}
	actionPrompter := newTerminalRMPrompter(strings.NewReader("2\n"), &bytes.Buffer{})
	action, cancelled := actionPrompter.SelectSmartAction(rmcommand.SmartActionPrompt{Message: "Select a clean action", Options: actions})
	if cancelled || action != actions[1] {
		t.Fatalf("smart action = (%#v, %t), want (%#v, false)", action, cancelled, actions[1])
	}

	targets := []string{"/tmp/one", "/tmp/two"}
	targetPrompter := newTerminalRMPrompter(strings.NewReader("\n"), &bytes.Buffer{})
	selected, cancelled := targetPrompter.SelectSmartTargets(rmcommand.SmartTargetPrompt{
		Message:       "Select items to delete",
		Options:       []rmcommand.SmartTargetChoice{{Value: targets[0], Label: "one"}, {Value: targets[1], Label: "two"}},
		InitialValues: targets,
	})
	if cancelled || strings.Join(selected, ",") != strings.Join(targets, ",") {
		t.Fatalf("default smart targets = (%#v, %t)", selected, cancelled)
	}

	nonePrompter := newTerminalRMPrompter(strings.NewReader("none\n"), &bytes.Buffer{})
	selected, cancelled = nonePrompter.SelectSmartTargets(rmcommand.SmartTargetPrompt{Options: []rmcommand.SmartTargetChoice{{Value: targets[0], Label: "one"}}})
	if cancelled || len(selected) != 0 {
		t.Fatalf("empty smart targets = (%#v, %t)", selected, cancelled)
	}
}

func TestTerminalRMPresenterWritesMappedMessages(t *testing.T) {
	output := &bytes.Buffer{}
	presenter := terminalRMPresenter{output: output}
	presenter.Intro("Remove")
	presenter.Paths([]string{"/tmp/one", "/tmp/two"})
	presenter.Notice("  not found, skipping: /tmp/missing")
	presenter.ProgressStart("Scanning...")
	presenter.ProgressStop("Found 1 target")
	presenter.Cancel("Cancelled.")
	presenter.Outro("Done!")

	want := "HACKYCY CLI\n\nRemove\n\n  /tmp/one\n  /tmp/two\n\n  not found, skipping: /tmp/missing\nScanning...\nFound 1 target\nCancelled.\nDone!\n"
	if output.String() != want {
		t.Fatalf("terminal rm output = %q, want %q", output.String(), want)
	}
}

func TestRMStandaloneBinary(t *testing.T) {
	repository := repositoryRoot(t)
	binary := standaloneBinaryOutputPath(filepath.Join(t.TempDir(), "ycy"))
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

	root := newStandaloneRMRoot(t)
	workingDirectory := filepath.Join(root, "project")
	if err := os.Mkdir(workingDirectory, 0o700); err != nil {
		t.Fatalf("create working directory: %v", err)
	}
	environment := environmentWith(map[string]string{"HOME": t.TempDir(), "USERPROFILE": ""})

	explicitFile := writeStandaloneRMFile(t, workingDirectory, "file.txt")
	explicitDirectory := filepath.Join(workingDirectory, "directory")
	if err := os.MkdirAll(filepath.Join(explicitDirectory, "nested"), 0o700); err != nil {
		t.Fatalf("create explicit directory: %v", err)
	}
	writeStandaloneRMFile(t, explicitDirectory, "nested/file.txt")
	output, err := runRMStandalone(binary, workingDirectory, environment, "yes\n", "rm", "file.txt", "directory", "missing")
	if err != nil || !strings.Contains(string(output), "not found, skipping:") || !strings.Contains(string(output), "Delete 2 items?") || !strings.Contains(string(output), "Deleted 2 items") || !strings.Contains(string(output), "Done!") {
		t.Fatalf("explicit rm = (%v, %q)", err, output)
	}
	for _, target := range []string{explicitFile, explicitDirectory} {
		if _, statErr := os.Lstat(target); !errors.Is(statErr, os.ErrNotExist) {
			t.Fatalf("explicit target %s = %v, want missing", target, statErr)
		}
	}

	cancelledTarget := writeStandaloneRMFile(t, workingDirectory, "cancelled.txt")
	output, err = runRMStandalone(binary, workingDirectory, environment, "\n", "rm", "cancelled.txt")
	if err != nil || !strings.Contains(string(output), "Cancelled.") {
		t.Fatalf("cancelled rm = (%v, %q)", err, output)
	}
	if _, statErr := os.Stat(cancelledTarget); statErr != nil {
		t.Fatalf("cancelled rm changed target: %v", statErr)
	}

	forcedTarget := writeStandaloneRMFile(t, workingDirectory, "forced.txt")
	output, err = runRMStandalone(binary, workingDirectory, environment, "", "rm", "--force", "forced.txt")
	if err != nil || strings.Contains(string(output), "Delete 1 item?") || !strings.Contains(string(output), "Done!") {
		t.Fatalf("forced rm = (%v, %q)", err, output)
	}
	if _, statErr := os.Stat(forcedTarget); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("forced target = %v, want missing", statErr)
	}

	output, err = runRMStandalone(binary, workingDirectory, environment, "", "rm", "still-missing")
	if err != nil || !strings.Contains(string(output), "No valid paths to delete.") {
		t.Fatalf("all missing rm = (%v, %q)", err, output)
	}

	smartDist := filepath.Join(workingDirectory, "dist")
	if err := os.MkdirAll(smartDist, 0o700); err != nil {
		t.Fatalf("create smart dist: %v", err)
	}
	writeStandaloneRMFile(t, smartDist, "artifact.txt")
	output, err = runRMStandalone(binary, workingDirectory, environment, "1\n\n", "rm")
	if err != nil || !strings.Contains(string(output), "Scanning...") || !strings.Contains(string(output), "Found 1 target") || !strings.Contains(string(output), "Done!") {
		t.Fatalf("smart rm = (%v, %q)", err, output)
	}
	if _, statErr := os.Stat(smartDist); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("smart dist = %v, want missing", statErr)
	}

	nodeModules := filepath.Join(workingDirectory, "node_modules")
	if err := os.MkdirAll(nodeModules, 0o700); err != nil {
		t.Fatalf("create node_modules: %v", err)
	}
	output, err = runRMStandalone(binary, workingDirectory, environment, "2\n", "rm", "--force")
	if err != nil || !strings.Contains(string(output), "Done!") {
		t.Fatalf("forced smart rm = (%v, %q)", err, output)
	}
	if _, statErr := os.Stat(nodeModules); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("forced smart node_modules = %v, want missing", statErr)
	}

	directDist := filepath.Join(workingDirectory, "dist")
	nestedDist := filepath.Join(workingDirectory, "packages", "app", "dist")
	if err := os.MkdirAll(directDist, 0o700); err != nil {
		t.Fatalf("create direct dist: %v", err)
	}
	if err := os.MkdirAll(nestedDist, 0o700); err != nil {
		t.Fatalf("create nested dist: %v", err)
	}
	output, err = runRMStandalone(binary, workingDirectory, environment, "3\n\n", "rm", "--depth", "0")
	if err != nil || !strings.Contains(string(output), "Found 1 target") {
		t.Fatalf("depth-zero smart rm = (%v, %q)", err, output)
	}
	if _, statErr := os.Stat(directDist); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("direct depth-zero dist = %v, want missing", statErr)
	}
	if _, statErr := os.Stat(nestedDist); statErr != nil {
		t.Fatalf("nested dist unexpectedly changed: %v", statErr)
	}

	helpOutput, err := runRMStandalone(binary, workingDirectory, environment, "", "--help")
	if err != nil || !strings.Contains(string(helpOutput), "rm") {
		t.Fatalf("root help = (%v, %q)", err, helpOutput)
	}
}

func runRMStandalone(binary, directory string, environment []string, input string, arguments ...string) ([]byte, error) {
	command := exec.Command(resolveStandaloneBinary(binary), arguments...)
	command.Dir = directory
	command.Env = environment
	command.Stdin = strings.NewReader(input)
	return command.CombinedOutput()
}

func writeStandaloneRMFile(t *testing.T, directory, name string) string {
	t.Helper()
	path := filepath.Join(directory, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("create parent for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte("contents"), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

func newStandaloneRMRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("get user home: %v", err)
	}
	for _, forbidden := range []string{workingDirectory, home, repositoryRoot(t)} {
		if standaloneRMPathsOverlap(root, forbidden) {
			t.Fatalf("disposable root %s overlaps forbidden path %s", root, forbidden)
		}
	}
	return root
}

func standaloneRMPathsOverlap(first, second string) bool {
	return standaloneRMPathContains(first, second) || standaloneRMPathContains(second, first)
}

func standaloneRMPathContains(root, path string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator)) && !filepath.IsAbs(relative)
}
