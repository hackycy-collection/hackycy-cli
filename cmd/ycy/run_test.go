package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/hackycy/hackycy-cli/internal/cliapp"
	runcommand "github.com/hackycy/hackycy-cli/internal/commands/run"
	"github.com/hackycy/hackycy-cli/internal/logging"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestTerminalRunAdapterTranslatesSelectionAndPresentation(t *testing.T) {
	experience := terminaltest.NewRecordingExperience(
		terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Value: "build"}},
		terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Value: string(runcommand.PackageManagerExternal)}},
	)
	run := experience.Open(context.Background())
	adapter := newTerminalRunAdapter(run, terminalexperience.Session{Kind: terminalexperience.RichInteractive, Color: true})

	script, cancelled, err := adapter.SelectScript(runcommand.ScriptPrompt{
		Message: "Select a script to run:",
		Options: []runcommand.ScriptChoice{
			{Value: "check", Label: "check", Hint: "go test ./..."},
			{Value: "build", Label: "build", Hint: "go build ./cmd/ycy"},
		},
	})
	if err != nil || cancelled || script != "build" {
		t.Fatalf("SelectScript() = (%q, %t, %v)", script, cancelled, err)
	}
	manager, cancelled, err := adapter.SelectPackageManager(runcommand.PackageManagerPrompt{
		Message: "Select a package manager:",
		Options: []runcommand.PackageManagerChoice{{Value: runcommand.PackageManagerExternal, Label: string(runcommand.PackageManagerExternal)}},
	})
	if err != nil || cancelled || manager != runcommand.PackageManagerExternal {
		t.Fatalf("SelectPackageManager() = (%q, %t, %v)", manager, cancelled, err)
	}
	adapter.Intro("Run Script")
	adapter.Info(string(runcommand.PackageManagerExternal) + " run build")
	adapter.Blank()
	adapter.Cancel("Operation cancelled.")
	if err := run.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	operations := experience.Run.Operations()
	if len(operations) != 7 || operations[0].Kind != terminaltest.AskOperation || operations[2].Kind != terminaltest.PresentOperation || operations[6].Kind != terminaltest.CloseOperation {
		t.Fatalf("operations = %#v", operations)
	}
	scriptRequest := operations[0].Value.(terminalexperience.InteractionRequest)
	if scriptRequest.Kind != terminalexperience.InteractionSelect || scriptRequest.Message != "Select a script to run:" || !reflect.DeepEqual(scriptRequest.Options, []terminalexperience.InteractionOption{
		{Label: "check", Value: "check", Description: "go test ./..."},
		{Label: "build", Value: "build", Description: "go build ./cmd/ycy"},
	}) || !reflect.DeepEqual(scriptRequest.CancelValues, []string{"", "q", "quit", "cancel"}) {
		t.Fatalf("script request = %#v", scriptRequest)
	}
	managerRequest := operations[1].Value.(terminalexperience.InteractionRequest)
	if managerRequest.Kind != terminalexperience.InteractionSelect || managerRequest.Message != "Select a package manager:" || !reflect.DeepEqual(managerRequest.Options, []terminalexperience.InteractionOption{{Label: string(runcommand.PackageManagerExternal), Value: string(runcommand.PackageManagerExternal)}}) {
		t.Fatalf("manager request = %#v", managerRequest)
	}
	intro := operations[2].Value.(terminalexperience.PresentationDocument)
	if !reflect.DeepEqual(intro.Blocks, []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleTitle, Text: "HACKYCY CLI"}, {Role: terminalexperience.VisualRoleActive, Text: "Run Script"}}) {
		t.Fatalf("intro document = %#v", intro)
	}
	if got := operations[5].Value.(terminalexperience.PresentationDocument).Blocks[0].Role; got != terminalexperience.VisualRoleWarning {
		t.Fatalf("cancellation role = %v", got)
	}
}

func TestTerminalRunAdapterMapsAutomationInteractionFailure(t *testing.T) {
	experience := terminaltest.NewRecordingExperience(terminaltest.SemanticAnswer{Err: terminalexperience.ErrAutomationInteraction})
	adapter := newTerminalRunAdapter(experience.Open(context.Background()), terminalexperience.Session{Kind: terminalexperience.Automation})
	if _, _, err := adapter.SelectScript(runcommand.ScriptPrompt{Options: []runcommand.ScriptChoice{{Value: "check", Label: "check"}}}); !errors.Is(err, errRunRequiresInteractive) {
		t.Fatalf("SelectScript() error = %v", err)
	}
}

func TestRunPlainSelectionReleasesFormBeforeRawChildIO(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("host shell fixture is Unix-specific")
	}
	root := t.TempDir()
	project := filepath.Join(root, "project")
	writeStandaloneRunFile(t, project, "package.json", `{"scripts":{"check":"echo check"}}`)
	writeStandaloneRunFile(t, project, "b"+"un"+".lock", "")
	binDirectory := filepath.Join(root, "bin")
	argumentsPath := filepath.Join(root, "arguments")
	workingDirectoryPath := filepath.Join(root, "working-directory")
	manager := filepath.Join(binDirectory, string(runcommand.PackageManagerExternal))
	managerScript := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$RUN_ARGUMENTS\"\npwd > \"$RUN_WORKING_DIRECTORY\"\nprintf 'child stdout'\nprintf 'child stderr' >&2\n"
	if err := os.MkdirAll(binDirectory, 0o700); err != nil {
		t.Fatalf("create manager directory: %v", err)
	}
	if err := os.WriteFile(manager, []byte(managerScript), 0o700); err != nil {
		t.Fatalf("write manager fixture: %v", err)
	}
	t.Setenv("PATH", binDirectory+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("RUN_ARGUMENTS", argumentsPath)
	t.Setenv("RUN_WORKING_DIRECTORY", workingDirectoryPath)
	withRunWorkingDirectory(t, project)

	rawInput := strings.NewReader("invalid\n1\n1\n")
	stdout, diagnostics, childStderr := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Session:     terminalexperience.Session{Kind: terminalexperience.PlainInteractive},
		Input:       rawInput,
		Output:      stdout,
		Diagnostics: diagnostics,
	})

	result, err := newRunHandler(experience, rawInput, stdout, childStderr)(context.Background(), runcommand.Input{})
	if err != nil || result.ExitCode != 0 {
		t.Fatalf("Run() = (%#v, %v)", result, err)
	}
	if got, want := stdout.String(), "HACKYCY CLI\n\nRun Script\n"+string(runcommand.PackageManagerExternal)+" run check\n\nchild stdout"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
	if !strings.Contains(diagnostics.String(), "Invalid selection") || !strings.Contains(diagnostics.String(), "check - echo check") || terminaltest.ContainsTerminalControl(diagnostics.Bytes()) {
		t.Fatalf("Plain diagnostics = %q", diagnostics.String())
	}
	if got, want := childStderr.String(), "child stderr"; got != want {
		t.Fatalf("child stderr = %q, want %q", got, want)
	}
	assertRunProcessFile(t, argumentsPath, "run\ncheck\n")
	resolvedProject, err := filepath.EvalSymlinks(project)
	if err != nil {
		t.Fatalf("resolve project path: %v", err)
	}
	assertRunProcessFile(t, workingDirectoryPath, resolvedProject+"\n")
}

func TestReleasedRunChildRunnerReleasesBeforeStartingChild(t *testing.T) {
	steps := []string{}
	runner := releasedRunChildRunner{
		release: func() error {
			steps = append(steps, "release")
			return nil
		},
		runner: runChildRunnerFunc(func(context.Context, runcommand.ChildRequest) (runcommand.Result, error) {
			steps = append(steps, "child")
			return runcommand.Result{}, nil
		}),
	}
	if _, err := runner.Run(context.Background(), runcommand.ChildRequest{}); err != nil || !reflect.DeepEqual(steps, []string{"release", "child"}) {
		t.Fatalf("Run() = (steps=%#v, err=%v)", steps, err)
	}
}

func TestRunAutomationFailsBeforeChildStartupOrInputRead(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "project")
	writeStandaloneRunFile(t, project, "package.json", `{"scripts":{"check":"echo check"}}`)
	startedPath := filepath.Join(root, "started")
	binDirectory := filepath.Join(root, "bin")
	manager := filepath.Join(binDirectory, string(runcommand.PackageManagerExternal))
	if err := os.MkdirAll(binDirectory, 0o700); err != nil {
		t.Fatalf("create manager directory: %v", err)
	}
	if err := os.WriteFile(manager, []byte("#!/bin/sh\ntouch \"$RUN_STARTED\"\n"), 0o700); err != nil {
		t.Fatalf("write manager fixture: %v", err)
	}
	t.Setenv("PATH", binDirectory+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("RUN_STARTED", startedPath)
	withRunWorkingDirectory(t, project)
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Session:     terminalexperience.Session{Kind: terminalexperience.Automation},
		Input:       panicRunReader{},
		Output:      stdout,
		Diagnostics: stderr,
	})
	app, err := cliapp.New(cliapp.BuildInfo{Version: "0.0.0-dev"}, cliapp.Dependencies{
		Out:     stdout,
		Err:     stderr,
		Logging: logging.NewRuntime(logging.Options{Writer: stderr}),
		Run:     newRunHandler(experience, panicRunReader{}, stdout, stderr),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	outcome := app.Execute(context.Background(), []string{"run"})
	if outcome.Code != 1 || !errors.Is(outcome.Err, errRunRequiresInteractive) || stdout.Len() != 0 || stderr.String() != "error: run requires an interactive terminal\n" {
		t.Fatalf("Automation outcome = %#v, streams = (%q, %q)", outcome, stdout.String(), stderr.String())
	}
	if terminaltest.ContainsTerminalControl(append(stdout.Bytes(), stderr.Bytes()...)) {
		t.Fatalf("Automation streams contain terminal control: (%q, %q)", stdout.String(), stderr.String())
	}
	if _, err := os.Stat(startedPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Automation failure launched child: %v", err)
	}
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

func withRunWorkingDirectory(t *testing.T, directory string) {
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

type runChildRunnerFunc func(context.Context, runcommand.ChildRequest) (runcommand.Result, error)

func (function runChildRunnerFunc) Run(ctx context.Context, request runcommand.ChildRequest) (runcommand.Result, error) {
	return function(ctx, request)
}

type panicRunReader struct{}

func (panicRunReader) Read([]byte) (int, error) {
	panic("run attempted to read Automation input")
}
