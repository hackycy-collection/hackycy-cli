package main

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/hackycy/hackycy-cli/internal/cliapp"
	zipcommand "github.com/hackycy/hackycy-cli/internal/commands/zip"
	"github.com/hackycy/hackycy-cli/internal/logging"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestTerminalZipAdapterTranslatesPlanningAndPresentation(t *testing.T) {
	experience := terminaltest.NewRecordingExperience(
		terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Value: "two"}},
		terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Value: "one"}},
		terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Values: []string{"**/*.html", "assets/**/*"}}},
		terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Value: "custom archive"}},
	)
	run := experience.Open(context.Background())
	adapter := newTerminalZipAdapter(run, terminalexperience.Session{Kind: terminalexperience.RichInteractive, Color: true})
	choices := []zipcommand.PlanningChoice{
		{Value: "one", Label: "one", Hint: "first"},
		{Value: "two", Label: "two", Hint: "second"},
	}
	packageRoot, cancelled, err := adapter.SelectPackage(zipcommand.SelectPackageStep{Message: "Select a package to zip:", Options: choices})
	if err != nil || cancelled || packageRoot != "two" {
		t.Fatalf("SelectPackage() = (%q, %t, %v)", packageRoot, cancelled, err)
	}
	source, cancelled, err := adapter.SelectSource(zipcommand.SelectSourceStep{Message: "Select a directory to zip:", Options: choices})
	if err != nil || cancelled || source != "one" {
		t.Fatalf("SelectSource() = (%q, %t, %v)", source, cancelled, err)
	}
	patterns, cancelled, err := adapter.SelectGlob(zipcommand.SelectGlobStep{
		Message:       "Select file patterns to include in the zip:",
		Options:       []zipcommand.PlanningChoice{{Value: "**/*", Label: "All"}, {Value: "**/*.html", Label: "HTML"}, {Value: "assets/**/*", Label: "Assets"}},
		InitialValues: []string{"**/*"},
	})
	if err != nil || cancelled || !reflect.DeepEqual(patterns, []string{"**/*.html", "assets/**/*"}) {
		t.Fatalf("SelectGlob() = (%#v, %t, %v)", patterns, cancelled, err)
	}
	filename, cancelled, err := adapter.EditOutputFile(zipcommand.EditOutputFileStep{Message: "Enter name", InitialValue: "default"})
	if err != nil || cancelled || filename != "custom archive" {
		t.Fatalf("EditOutputFile() = (%q, %t, %v)", filename, cancelled, err)
	}
	adapter.Intro()
	adapter.Note(zipcommand.PlanningNote{Title: "Zip plan", Lines: []string{"Source: dist", "Output: archive.zip"}})
	adapter.Progress("Collecting files...")
	adapter.Cancel("Operation cancelled.")
	adapter.Outro("Done!")
	if err := run.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	operations := experience.Run.Operations()
	if len(operations) != 10 || operations[0].Kind != terminaltest.AskOperation || operations[4].Kind != terminaltest.PresentOperation || operations[9].Kind != terminaltest.CloseOperation {
		t.Fatalf("operations = %#v", operations)
	}
	packageRequest := operations[0].Value.(terminalexperience.InteractionRequest)
	if packageRequest.Kind != terminalexperience.InteractionSelect || packageRequest.Message != "Select a package to zip:" || !packageRequest.HasDefault || packageRequest.Default.Value != "one" || !reflect.DeepEqual(packageRequest.Options, []terminalexperience.InteractionOption{{Label: "one", Value: "one", Description: "first"}, {Label: "two", Value: "two", Description: "second"}}) || !reflect.DeepEqual(packageRequest.CancelValues, []string{"q", "quit", "cancel"}) || packageRequest.PlainLead != "Select a package to zip:" || packageRequest.PlainPrompt != "> " || packageRequest.ParsePlain == nil {
		t.Fatalf("package request = %#v", packageRequest)
	}
	globRequest := operations[2].Value.(terminalexperience.InteractionRequest)
	if globRequest.Kind != terminalexperience.InteractionMultiSelect || !globRequest.HasDefault || !reflect.DeepEqual(globRequest.Default.Values, []string{"**/*"}) || globRequest.ParsePlain == nil {
		t.Fatalf("glob request = %#v", globRequest)
	}
	outputRequest := operations[3].Value.(terminalexperience.InteractionRequest)
	if outputRequest.Kind != terminalexperience.InteractionText || outputRequest.Placeholder != "default" || !outputRequest.HasDefault || outputRequest.Default.Value != "default" || outputRequest.PlainPrompt != "Enter name [default]: " || outputRequest.ParsePlain == nil {
		t.Fatalf("output request = %#v", outputRequest)
	}
	intro := operations[4].Value.(terminalexperience.PresentationDocument)
	if !reflect.DeepEqual(intro.Blocks, []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleTitle, Text: "HACKYCY CLI"}, {Role: terminalexperience.VisualRoleActive, Text: "Zip Directory"}}) {
		t.Fatalf("intro document = %#v", intro)
	}
	note := operations[5].Value.(terminalexperience.PresentationDocument)
	if !reflect.DeepEqual(note.Blocks, []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleActive, Text: "Zip plan"}, {Role: terminalexperience.VisualRoleMuted, Text: "Source: dist\nOutput: archive.zip"}}) {
		t.Fatalf("note document = %#v", note)
	}
	if got := operations[8].Value.(terminalexperience.PresentationDocument).Blocks[0].Role; got != terminalexperience.VisualRoleSuccess {
		t.Fatalf("success role = %v", got)
	}
}

func TestTerminalZipAdapterPlainPreservesLegacyInputGrammar(t *testing.T) {
	stdout, diagnostics := &bytes.Buffer{}, &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Session:     terminalexperience.Session{Kind: terminalexperience.PlainInteractive},
		Input:       strings.NewReader("invalid\n2\n\n2,3\n\nall\nnone\n\n custom archive \ncancel\n"),
		Output:      stdout,
		Diagnostics: diagnostics,
	})
	run := experience.Open(context.Background())
	adapter := newTerminalZipAdapter(run, experience.Session())
	choices := []zipcommand.PlanningChoice{{Value: "one", Label: "one", Hint: "first"}, {Value: "two", Label: "two", Hint: "second"}, {Value: "three", Label: "three"}}
	packageRoot, cancelled, err := adapter.SelectPackage(zipcommand.SelectPackageStep{Message: "Select a package to zip:", Options: choices})
	if err != nil || cancelled || packageRoot != "two" {
		t.Fatalf("SelectPackage() = (%q, %t, %v)", packageRoot, cancelled, err)
	}
	source, cancelled, err := adapter.SelectSource(zipcommand.SelectSourceStep{Message: "Select a directory to zip:", Options: choices})
	if err != nil || cancelled || source != "one" {
		t.Fatalf("SelectSource() = (%q, %t, %v)", source, cancelled, err)
	}
	globStep := zipcommand.SelectGlobStep{Message: "Select file patterns to include in the zip:", Options: []zipcommand.PlanningChoice{{Value: "**/*", Label: "All"}, {Value: "**/*.html", Label: "HTML"}, {Value: "assets/**/*", Label: "Assets"}}, InitialValues: []string{"**/*"}}
	patterns, cancelled, err := adapter.SelectGlob(globStep)
	if err != nil || cancelled || !reflect.DeepEqual(patterns, []string{"**/*.html", "assets/**/*"}) {
		t.Fatalf("SelectGlob() = (%#v, %t, %v)", patterns, cancelled, err)
	}
	for _, value := range [][]string{{"**/*"}, {"**/*"}, {"**/*"}} {
		patterns, cancelled, err = adapter.SelectGlob(globStep)
		if err != nil || cancelled || !reflect.DeepEqual(patterns, value) {
			t.Fatalf("default glob = (%#v, %t, %v), want %#v", patterns, cancelled, err, value)
		}
	}
	name, cancelled, err := adapter.EditOutputFile(zipcommand.EditOutputFileStep{Message: "Enter name", InitialValue: "default"})
	if err != nil || cancelled || name != "default" {
		t.Fatalf("EditOutputFile() default = (%q, %t, %v)", name, cancelled, err)
	}
	name, cancelled, err = adapter.EditOutputFile(zipcommand.EditOutputFileStep{Message: "Enter name", InitialValue: "default"})
	if err != nil || cancelled || name != "custom archive" {
		t.Fatalf("EditOutputFile() = (%q, %t, %v)", name, cancelled, err)
	}
	_, cancelled, err = adapter.SelectSource(zipcommand.SelectSourceStep{Message: "Select a directory to zip:", Options: choices})
	if err != nil || !cancelled {
		t.Fatalf("SelectSource() cancellation = (%t, %v)", cancelled, err)
	}
	if stdout.Len() != 0 || !strings.Contains(diagnostics.String(), "Invalid selection") || !strings.Contains(diagnostics.String(), "2) two - second") || !strings.Contains(diagnostics.String(), "Enter name [default]: ") || terminaltest.ContainsTerminalControl(append(stdout.Bytes(), diagnostics.Bytes()...)) {
		t.Fatalf("Plain streams = (%q, %q)", stdout.String(), diagnostics.String())
	}
	if err := run.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	for _, cancellation := range []string{"q", "quit", "cancel"} {
		t.Run(cancellation, func(t *testing.T) {
			experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
				Session:     terminalexperience.Session{Kind: terminalexperience.PlainInteractive},
				Input:       strings.NewReader(cancellation + "\n"),
				Output:      io.Discard,
				Diagnostics: io.Discard,
			})
			run := experience.Open(context.Background())
			_, cancelled, err := newTerminalZipAdapter(run, experience.Session()).SelectSource(zipcommand.SelectSourceStep{Options: choices})
			if err != nil || !cancelled {
				t.Fatalf("SelectSource() cancellation = (%t, %v)", cancelled, err)
			}
			if err := run.Close(); err != nil {
				t.Fatalf("Close() error = %v", err)
			}
		})
	}
}

func TestZIPPlainJourneyCreatesAStructuralArchive(t *testing.T) {
	project := t.TempDir()
	writeStandaloneZIPFile(t, project, "package.json", `{"name":"project","devDependencies":{"vite":"1"}}`)
	writeStandaloneZIPFile(t, project, "dist/index.html", "<main />")
	writeStandaloneZIPFile(t, project, "dist/assets/app.js", "console.log('app')")
	writeStandaloneZIPFile(t, project, "dist/.secret", "not archived")
	withZIPWorkingDirectory(t, project)
	stdout, diagnostics := &bytes.Buffer{}, &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Session:     terminalexperience.Session{Kind: terminalexperience.PlainInteractive},
		Input:       strings.NewReader("\n\n\n"),
		Output:      stdout,
		Diagnostics: diagnostics,
	})

	result, err := newZipHandler(experience)(context.Background(), zipcommand.Input{Directory: ".", Open: false, WithDir: "bundle"})
	if err != nil || result.Kind != zipcommand.ResultCompleted || !strings.Contains(stdout.String(), "Zip Directory") || !strings.Contains(stdout.String(), "Done!") || terminaltest.ContainsTerminalControl(append(stdout.Bytes(), diagnostics.Bytes()...)) {
		t.Fatalf("Plain zip = (%#v, %v), streams = (%q, %q)", result, err, stdout.String(), diagnostics.String())
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
	if err := archive.Close(); err != nil {
		t.Fatalf("close archive: %v", err)
	}
	want := map[string]string{"bundle/index.html": "<main />", "bundle/assets/app.js": "console.log('app')"}
	if !reflect.DeepEqual(contents, want) {
		t.Fatalf("archive contents = %#v, want %#v", contents, want)
	}
}

func TestZIPAutomationFailsBeforeArchiveCreationOrInputRead(t *testing.T) {
	project := t.TempDir()
	writeStandaloneZIPFile(t, project, "package.json", `{"name":"project","devDependencies":{"vite":"1"}}`)
	writeStandaloneZIPFile(t, project, "dist/index.html", "<main />")
	withZIPWorkingDirectory(t, project)
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Session:     terminalexperience.Session{Kind: terminalexperience.Automation},
		Input:       panicZipReader{},
		Output:      stdout,
		Diagnostics: stderr,
	})
	app, err := cliapp.New(cliapp.BuildInfo{Version: "0.0.0-dev"}, cliapp.Dependencies{
		Out:     stdout,
		Err:     stderr,
		Logging: logging.NewRuntime(logging.Options{Writer: stderr}),
		ZIP:     newZipHandler(experience),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	outcome := app.Execute(context.Background(), []string{"zip", ".", "--without-open"})
	if outcome.Code != 1 || !errors.Is(outcome.Err, errZipRequiresInteractive) || stdout.Len() != 0 || stderr.String() != "error: zip requires an interactive terminal\n" || terminaltest.ContainsTerminalControl(append(stdout.Bytes(), stderr.Bytes()...)) {
		t.Fatalf("Automation outcome = %#v, streams = (%q, %q)", outcome, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(filepath.Join(project, "dist", "project.zip")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Automation created archive: %v", err)
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

func TestZIPStandaloneBinaryRejectsRedirectedPlanningAndPreservesHelp(t *testing.T) {
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
	environment := environmentWith(map[string]string{"HOME": t.TempDir(), "USERPROFILE": ""})
	command := exec.Command(resolveStandaloneBinary(binary), "zip", ".", "--without-open", "--with-dir", "bundle")
	command.Dir = project
	command.Env = environment
	command.Stdin = strings.NewReader("\n\n\n")
	output, err := command.CombinedOutput()
	if err == nil || string(output) != "error: zip requires an interactive terminal\n" {
		t.Fatalf("redirected zip = (%v, %q)", err, output)
	}
	archivePath := filepath.Join(project, "dist", "project.zip")
	if _, err := os.Stat(archivePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("redirected zip created archive: %v", err)
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

type panicZipReader struct{}

func (panicZipReader) Read([]byte) (int, error) {
	panic("zip Automation must not read stdin")
}

func withZIPWorkingDirectory(t *testing.T, directory string) {
	t.Helper()
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir(directory); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(workingDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
}
