package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	zipcommand "github.com/hackycy/hackycy-cli/internal/commands/zip"
)

func TestTerminalZipPrompterSupportsDefaultsSelectionsAndCancellation(t *testing.T) {
	output := &bytes.Buffer{}
	prompter := newTerminalZipPrompter(strings.NewReader("2\n\n2,3\ncustom archive\n"), output)
	choices := []zipcommand.PlanningChoice{
		{Value: "one", Label: "one", Hint: "first"},
		{Value: "two", Label: "two", Hint: "second"},
	}
	packageRoot, cancelled := prompter.SelectPackage(zipcommand.SelectPackageStep{Message: "Select a package to zip:", Options: choices})
	if cancelled || packageRoot != "two" {
		t.Fatalf("SelectPackage() = (%q, %t)", packageRoot, cancelled)
	}
	source, cancelled := prompter.SelectSource(zipcommand.SelectSourceStep{Message: "Select a directory to zip:", Options: choices})
	if cancelled || source != "one" {
		t.Fatalf("SelectSource() = (%q, %t)", source, cancelled)
	}
	patterns, cancelled := prompter.SelectGlob(zipcommand.SelectGlobStep{
		Message:       "Select file patterns to include in the zip:",
		Options:       []zipcommand.PlanningChoice{{Value: "**/*", Label: "All"}, {Value: "**/*.html", Label: "HTML"}, {Value: "assets/**/*", Label: "Assets"}},
		InitialValues: []string{"**/*"},
	})
	if cancelled || !reflect.DeepEqual(patterns, []string{"**/*.html", "assets/**/*"}) {
		t.Fatalf("SelectGlob() = (%#v, %t)", patterns, cancelled)
	}
	filename, cancelled := prompter.EditOutputFile(zipcommand.EditOutputFileStep{Message: "Enter name", InitialValue: "default"})
	if cancelled || filename != "custom archive" {
		t.Fatalf("EditOutputFile() = (%q, %t)", filename, cancelled)
	}
	if !strings.Contains(output.String(), "2) two - second") || !strings.Contains(output.String(), "Select file patterns") {
		t.Fatalf("prompt output = %q", output.String())
	}

	cancelledPrompter := newTerminalZipPrompter(strings.NewReader("cancel\n"), &bytes.Buffer{})
	_, cancelled = cancelledPrompter.SelectSource(zipcommand.SelectSourceStep{Options: choices})
	if !cancelled {
		t.Fatal("SelectSource() did not treat cancel as cancellation")
	}
}

func TestTerminalZipPrompterPreservesDefaultGlobAndOutputName(t *testing.T) {
	prompter := newTerminalZipPrompter(strings.NewReader("\n\n"), &bytes.Buffer{})
	patterns, cancelled := prompter.SelectGlob(zipcommand.SelectGlobStep{Options: []zipcommand.PlanningChoice{{Value: "**/*"}}, InitialValues: []string{"**/*"}})
	if cancelled || !reflect.DeepEqual(patterns, []string{"**/*"}) {
		t.Fatalf("SelectGlob() = (%#v, %t)", patterns, cancelled)
	}
	name, cancelled := prompter.EditOutputFile(zipcommand.EditOutputFileStep{InitialValue: "default"})
	if cancelled || name != "default" {
		t.Fatalf("EditOutputFile() = (%q, %t)", name, cancelled)
	}
}

func TestTerminalZipPresenterWritesMappedMessages(t *testing.T) {
	output := &bytes.Buffer{}
	presenter := terminalZipPresenter{output: output}
	presenter.Intro()
	presenter.Note(zipcommand.PlanningNote{Title: "Zip plan", Lines: []string{"Source: dist", "Output: archive.zip"}})
	presenter.Progress("Collecting files...")
	presenter.Cancel("Operation cancelled.")
	presenter.Outro("Done!")
	want := "HACKYCY CLI\n\nZip Directory\nZip plan\nSource: dist\nOutput: archive.zip\nCollecting files...\nOperation cancelled.\nDone!\n"
	if output.String() != want {
		t.Fatalf("presentation = %q, want %q", output.String(), want)
	}
}

func TestZipRemoteNameResolverPrefersOriginAndIgnoresUnusableOutput(t *testing.T) {
	resolver := newZipRemoteNameResolver(zipRemoteOutputRunnerFunc(func(directory string) ([]byte, error) {
		if directory != "/workspace" {
			t.Fatalf("directory = %q", directory)
		}
		return []byte("upstream\thttps://github.com/example/upstream.git (fetch)\norigin\tgit@github.com:example/project.git (fetch)\n"), nil
	}))
	name, err := resolver.ResolveRemoteName("/workspace")
	if err != nil || name != "example-project" {
		t.Fatalf("ResolveRemoteName() = (%q, %v)", name, err)
	}

	resolver = newZipRemoteNameResolver(zipRemoteOutputRunnerFunc(func(string) ([]byte, error) {
		return []byte("origin\tinvalid remote (fetch)\n"), nil
	}))
	name, err = resolver.ResolveRemoteName("/workspace")
	if err != nil || name != "" {
		t.Fatalf("invalid remote = (%q, %v)", name, err)
	}

	wantError := errors.New("git unavailable")
	resolver = newZipRemoteNameResolver(zipRemoteOutputRunnerFunc(func(string) ([]byte, error) {
		return nil, wantError
	}))
	_, err = resolver.ResolveRemoteName("/workspace")
	if !errors.Is(err, wantError) {
		t.Fatalf("resolver error = %v, want %v", err, wantError)
	}
}

func TestOSZipRemoteOutputRunnerReadsADisposableRepository(t *testing.T) {
	repository := t.TempDir()
	for _, arguments := range [][]string{{"init"}, {"remote", "add", "origin", "https://github.com/example/project.git"}} {
		command := exec.Command("git", arguments...)
		command.Dir = repository
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("git %q: %v\n%s", arguments, err, output)
		}
	}
	resolver := newZipRemoteNameResolver(osZipRemoteOutputRunner{})
	name, err := resolver.ResolveRemoteName(repository)
	if err != nil || name != "example-project" {
		t.Fatalf("ResolveRemoteName() = (%q, %v)", name, err)
	}
}

func TestHostZipRevealerUsesPlatformCommands(t *testing.T) {
	testCases := []struct {
		goos string
		name string
		args []string
	}{
		{goos: "darwin", name: "open", args: []string{"/tmp/archive.zip"}},
		{goos: "linux", name: "xdg-open", args: []string{"/tmp/archive.zip"}},
		{goos: "windows", name: "cmd", args: []string{"/c", "start", "", "/tmp/archive.zip"}},
	}
	for _, testCase := range testCases {
		t.Run(testCase.goos, func(t *testing.T) {
			name, args, err := zipRevealCommand(testCase.goos, "/tmp/archive.zip")
			if err != nil || name != testCase.name || !reflect.DeepEqual(args, testCase.args) {
				t.Fatalf("zipRevealCommand() = (%q, %#v, %v)", name, args, err)
			}
		})
	}
	if _, _, err := zipRevealCommand("plan9", "/tmp/archive.zip"); err == nil {
		t.Fatal("unsupported reveal platform did not fail")
	}

	runner := &recordingZipHostRunner{}
	revealer := newHostZipRevealer(runner)
	if err := revealer.Reveal("/tmp/archive.zip"); err != nil {
		t.Fatalf("Reveal() error = %v", err)
	}
	wantName, wantArgs, _ := zipRevealCommand(runtime.GOOS, "/tmp/archive.zip")
	if runner.name != wantName || !reflect.DeepEqual(runner.arguments, wantArgs) {
		t.Fatalf("runner = %#v, want (%q, %#v)", runner, wantName, wantArgs)
	}
}

func TestZIPStandaloneBinaryCreatesAStructuralArchive(t *testing.T) {
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

	project := filepath.Join(t.TempDir(), "project")
	writeStandaloneZIPFile(t, project, "package.json", `{"name":"project","devDependencies":{"vite":"1"}}`)
	writeStandaloneZIPFile(t, project, "dist/index.html", "<main />")
	writeStandaloneZIPFile(t, project, "dist/assets/app.js", "console.log('app')")
	writeStandaloneZIPFile(t, project, "dist/.secret", "not archived")
	environment := environmentWith(map[string]string{"HOME": t.TempDir(), "USERPROFILE": ""})
	command := exec.Command(resolveStandaloneBinary(binary), "zip", ".", "--without-open", "--with-dir", "bundle")
	command.Dir = project
	command.Env = environment
	command.Stdin = strings.NewReader("\n\n\n")
	output, err := command.CombinedOutput()
	if err != nil || !strings.Contains(string(output), "Zip Directory") || !strings.Contains(string(output), "Done!") {
		t.Fatalf("zip standalone = (%v, %q)", err, output)
	}
	archivePath := filepath.Join(project, "dist", "project.zip")
	archive, err := zip.OpenReader(archivePath)
	if err != nil {
		t.Fatalf("open archive: %v", err)
	}
	contents := make(map[string]string)
	for _, file := range archive.File {
		reader, err := file.Open()
		if err != nil {
			t.Fatalf("open archive entry: %v", err)
		}
		bytes, err := io.ReadAll(reader)
		_ = reader.Close()
		if err != nil {
			t.Fatalf("read archive entry: %v", err)
		}
		contents[file.Name] = string(bytes)
	}
	want := map[string]string{"bundle/index.html": "<main />", "bundle/assets/app.js": "console.log('app')"}
	if !reflect.DeepEqual(contents, want) {
		t.Fatalf("archive contents = %#v, want %#v", contents, want)
	}
	if err := archive.Close(); err != nil {
		t.Fatalf("close archive: %v", err)
	}

	if err := os.Remove(archivePath); err != nil {
		t.Fatalf("remove archive before cancellation: %v", err)
	}
	command = exec.Command(resolveStandaloneBinary(binary), "zip", ".", "--without-open")
	command.Dir = project
	command.Env = environment
	command.Stdin = strings.NewReader("cancel\n")
	output, err = command.CombinedOutput()
	if err != nil || !strings.Contains(string(output), "Operation cancelled.") {
		t.Fatalf("zip cancellation = (%v, %q)", err, output)
	}
	if _, err := os.Stat(archivePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("cancellation archive = %v, want missing", err)
	}

	command = exec.Command(resolveStandaloneBinary(binary), "zip", "--help")
	command.Dir = project
	command.Env = environment
	output, err = command.CombinedOutput()
	if err != nil || !strings.Contains(string(output), "Zip a directory into a zip file") || !strings.Contains(string(output), "--without-open") {
		t.Fatalf("zip help = (%v, %q)", err, output)
	}
}

type zipRemoteOutputRunnerFunc func(string) ([]byte, error)

func (function zipRemoteOutputRunnerFunc) Output(directory string) ([]byte, error) {
	return function(directory)
}

type recordingZipHostRunner struct {
	name      string
	arguments []string
}

func (runner *recordingZipHostRunner) Run(name string, arguments []string) error {
	runner.name = name
	runner.arguments = append([]string(nil), arguments...)
	return nil
}

func writeStandaloneZIPFile(t *testing.T, root, name, contents string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("create parent for %s: %v", name, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}
