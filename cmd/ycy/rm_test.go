package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/hackycy/hackycy-cli/internal/cliapp"
	rmcommand "github.com/hackycy/hackycy-cli/internal/commands/rm"
	"github.com/hackycy/hackycy-cli/internal/logging"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestTerminalRMAdapterTranslatesPromptClusterAndPresentation(t *testing.T) {
	experience := terminaltest.NewRecordingExperience(
		terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Confirmed: true}},
		terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Value: "node-dist"}},
		terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Values: []string{"/project/dist"}}},
	)
	run := experience.Open(context.Background())
	adapter := newTerminalRMAdapter(run, terminalexperience.Session{Kind: terminalexperience.RichInteractive, Color: true})

	confirmed, cancelled, err := adapter.ConfirmExplicit(rmcommand.ExplicitConfirmationPrompt{Message: "Delete 1 item?"})
	if err != nil || cancelled || !confirmed {
		t.Fatalf("ConfirmExplicit() = (%t, %t, %v)", confirmed, cancelled, err)
	}
	action, cancelled, err := adapter.SelectSmartAction(rmcommand.SmartActionPrompt{Message: "Select a clean action", Options: []rmcommand.SmartAction{{ID: "node-dist", Label: "Node project - delete ./dist"}}})
	if err != nil || cancelled || action.ID != "node-dist" {
		t.Fatalf("SelectSmartAction() = (%#v, %t, %v)", action, cancelled, err)
	}
	targets, cancelled, err := adapter.SelectSmartTargets(rmcommand.SmartTargetPrompt{
		Message:       "Select items to delete",
		Options:       []rmcommand.SmartTargetChoice{{Value: "/project/dist", Label: "dist"}},
		InitialValues: []string{"/project/dist"},
	})
	if err != nil || cancelled || !reflect.DeepEqual(targets, []string{"/project/dist"}) {
		t.Fatalf("SelectSmartTargets() = (%#v, %t, %v)", targets, cancelled, err)
	}
	adapter.Intro("Remove")
	adapter.Paths([]string{"/project/dist"})
	adapter.Notice("  not found, skipping: /project/missing")
	adapter.ProgressStart("Scanning...")
	adapter.ProgressStop("Found 1 target")
	adapter.Cancel("Cancelled.")
	adapter.Outro("Done!")
	if err := run.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	operations := experience.Run.Operations()
	if len(operations) != 11 || operations[0].Kind != terminaltest.AskOperation || operations[3].Kind != terminaltest.PresentOperation || operations[10].Kind != terminaltest.CloseOperation {
		t.Fatalf("operations = %#v", operations)
	}
	confirm := operations[0].Value.(terminalexperience.InteractionRequest)
	if confirm.Kind != terminalexperience.InteractionConfirm || confirm.Message != "Delete 1 item?" || !confirm.HasDefault || confirm.Default.Confirmed {
		t.Fatalf("confirmation request = %#v", confirm)
	}
	actionRequest := operations[1].Value.(terminalexperience.InteractionRequest)
	if actionRequest.Kind != terminalexperience.InteractionSelect || actionRequest.Message != "Select a clean action" || !actionRequest.HasDefault || actionRequest.Default.Value != "node-dist" || !reflect.DeepEqual(actionRequest.CancelValues, []string{"q", "quit", "cancel"}) {
		t.Fatalf("action request = %#v", actionRequest)
	}
	targetRequest := operations[2].Value.(terminalexperience.InteractionRequest)
	if targetRequest.Kind != terminalexperience.InteractionMultiSelect || targetRequest.Message != "Select items to delete" || !targetRequest.HasDefault || !reflect.DeepEqual(targetRequest.Default.Values, []string{"/project/dist"}) {
		t.Fatalf("target request = %#v", targetRequest)
	}
	intro := operations[3].Value.(terminalexperience.PresentationDocument)
	if !reflect.DeepEqual(intro.Blocks, []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleTitle, Text: "HACKYCY CLI"}, {Role: terminalexperience.VisualRoleActive, Text: "Remove"}}) {
		t.Fatalf("intro document = %#v", intro)
	}
	if got := operations[9].Value.(terminalexperience.PresentationDocument).Blocks[0].Role; got != terminalexperience.VisualRoleSuccess {
		t.Fatalf("success role = %v", got)
	}
}

func TestTerminalRMAdapterPlainPreservesLegacyInputGrammarAndMutationBoundaries(t *testing.T) {
	root := t.TempDir()
	confirmed := writeStandaloneRMFile(t, root, "confirmed.txt")
	cancelledTarget := writeStandaloneRMFile(t, root, "cancelled.txt")

	for _, testCase := range []struct {
		name       string
		input      string
		target     string
		shouldGone bool
		contains   []string
	}{
		{name: "confirmed", input: "maybe\nyes\n", target: confirmed, shouldGone: true, contains: []string{"Invalid confirmation", "Delete 1 item? [y/N]:", "Deleted 1 item", "Done!"}},
		{name: "declined", input: "\n", target: cancelledTarget, contains: []string{"Delete 1 item? [y/N]:", "Cancelled."}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			stdout, diagnostics := &bytes.Buffer{}, &bytes.Buffer{}
			experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
				Session:     terminalexperience.Session{Kind: terminalexperience.PlainInteractive},
				Input:       strings.NewReader(testCase.input),
				Output:      stdout,
				Diagnostics: diagnostics,
			})
			withRMWorkingDirectory(t, root)

			result, err := newRMHandler(experience)(context.Background(), rmcommand.Input{Paths: []string{filepath.Base(testCase.target)}})
			if err != nil || result != (rmcommand.Result{}) {
				t.Fatalf("Run() = (%#v, %v)", result, err)
			}
			allOutput := append(append([]byte{}, stdout.Bytes()...), diagnostics.Bytes()...)
			if terminaltest.ContainsTerminalControl(allOutput) {
				t.Fatalf("Plain streams contain terminal control: (%q, %q)", stdout.String(), diagnostics.String())
			}
			for _, want := range testCase.contains {
				if !strings.Contains(stdout.String()+diagnostics.String(), want) {
					t.Fatalf("streams = (%q, %q), missing %q", stdout.String(), diagnostics.String(), want)
				}
			}
			_, statErr := os.Stat(testCase.target)
			if testCase.shouldGone && !errors.Is(statErr, os.ErrNotExist) {
				t.Fatalf("confirmed target = %v, want missing", statErr)
			}
			if !testCase.shouldGone && statErr != nil {
				t.Fatalf("declined target = %v, want retained", statErr)
			}
		})
	}
}

func TestRMAutomationPreservesForceAndNoTargetPathsAndFailsPromptPathsBeforeEffects(t *testing.T) {
	root := t.TempDir()
	forcedTarget := writeStandaloneRMFile(t, root, "forced.txt")
	promptTarget := writeStandaloneRMFile(t, root, "prompt.txt")
	withRMWorkingDirectory(t, root)
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	for _, testCase := range []struct {
		name      string
		input     rmcommand.Input
		wantOut   string
		wantError error
	}{
		{name: "force explicit", input: rmcommand.Input{Paths: []string{"forced.txt"}, Force: true}, wantOut: "HACKYCY CLI\n\nRemove\nDeleting 1 item...\nDeleted 1 item\nDone!\n"},
		{name: "missing explicit", input: rmcommand.Input{Paths: []string{"missing.txt"}}, wantOut: "HACKYCY CLI\n\nRemove\n  not found, skipping: " + filepath.Join(workingDirectory, "missing.txt") + "\nNo valid paths to delete.\n"},
		{name: "explicit confirmation", input: rmcommand.Input{Paths: []string{"prompt.txt"}}, wantError: errRMRequiresInteractive},
		{name: "smart selection", input: rmcommand.Input{}, wantError: errRMRequiresInteractive},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			stdout, diagnostics := &bytes.Buffer{}, &bytes.Buffer{}
			experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
				Session:     terminalexperience.Session{Kind: terminalexperience.Automation},
				Input:       panicRMReader{},
				Output:      stdout,
				Diagnostics: diagnostics,
			})
			result, err := newRMHandler(experience)(context.Background(), testCase.input)
			if testCase.wantError == nil {
				if err != nil || result != (rmcommand.Result{}) || stdout.String() != testCase.wantOut || diagnostics.Len() != 0 {
					t.Fatalf("Automation result = (%#v, %v), streams = (%q, %q)", result, err, stdout.String(), diagnostics.String())
				}
				return
			}
			if !errors.Is(err, testCase.wantError) || result != (rmcommand.Result{}) || stdout.Len() != 0 || diagnostics.Len() != 0 {
				t.Fatalf("Automation failure = (%#v, %v), streams = (%q, %q)", result, err, stdout.String(), diagnostics.String())
			}
		})
	}
	if _, err := os.Stat(promptTarget); err != nil {
		t.Fatalf("Automation prompt failure changed target: %v", err)
	}
	if _, err := os.Stat(forcedTarget); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("forced target = %v, want missing", err)
	}
}

func TestRMAutomationErrorUsesStderrWithoutPartialCommandResult(t *testing.T) {
	root := t.TempDir()
	target := writeStandaloneRMFile(t, root, "target.txt")
	withRMWorkingDirectory(t, root)
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Session:     terminalexperience.Session{Kind: terminalexperience.Automation},
		Input:       panicRMReader{},
		Output:      stdout,
		Diagnostics: stderr,
	})
	app, err := cliapp.New(cliapp.BuildInfo{Version: "0.0.0-dev"}, cliapp.Dependencies{
		Out:     stdout,
		Err:     stderr,
		Logging: logging.NewRuntime(logging.Options{Writer: stderr}),
		RM:      newRMHandler(experience),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	outcome := app.Execute(context.Background(), []string{"rm", filepath.Base(target)})
	if outcome.Code != 1 || !errors.Is(outcome.Err, errRMRequiresInteractive) || stdout.Len() != 0 || stderr.String() != "error: rm requires an interactive terminal\n" {
		t.Fatalf("Automation outcome = %#v, streams = (%q, %q)", outcome, stdout.String(), stderr.String())
	}
	if terminaltest.ContainsTerminalControl(append(stdout.Bytes(), stderr.Bytes()...)) {
		t.Fatalf("Automation streams contain terminal control: (%q, %q)", stdout.String(), stderr.String())
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("Automation failure changed target: %v", err)
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
	output, err := runRMStandalone(binary, workingDirectory, environment, "", "rm", "--force", "file.txt", "directory", "missing")
	if err != nil || !strings.Contains(string(output), "not found, skipping:") || strings.Contains(string(output), "Delete 2 items?") || !strings.Contains(string(output), "Deleted 2 items") || !strings.Contains(string(output), "Done!") {
		t.Fatalf("explicit rm = (%v, %q)", err, output)
	}
	for _, target := range []string{explicitFile, explicitDirectory} {
		if _, statErr := os.Lstat(target); !errors.Is(statErr, os.ErrNotExist) {
			t.Fatalf("explicit target %s = %v, want missing", target, statErr)
		}
	}

	cancelledTarget := writeStandaloneRMFile(t, workingDirectory, "cancelled.txt")
	output, err = runRMStandalone(binary, workingDirectory, environment, "\n", "rm", "cancelled.txt")
	if err == nil || string(output) != "error: rm requires an interactive terminal\n" {
		t.Fatalf("redirected confirmation rm = (%v, %q)", err, output)
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
		t.Fatalf("redirected no-target rm = (%v, %q)", err, output)
	}

	smartDist := filepath.Join(workingDirectory, "dist")
	if err := os.MkdirAll(smartDist, 0o700); err != nil {
		t.Fatalf("create smart dist: %v", err)
	}
	output, err = runRMStandalone(binary, workingDirectory, environment, "1\n", "rm")
	if err == nil || string(output) != "error: rm requires an interactive terminal\n" {
		t.Fatalf("redirected smart rm = (%v, %q)", err, output)
	}
	if _, statErr := os.Stat(smartDist); statErr != nil {
		t.Fatalf("redirected smart rm changed dist: %v", statErr)
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

func withRMWorkingDirectory(t *testing.T, directory string) {
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

type panicRMReader struct{}

func (panicRMReader) Read([]byte) (int, error) {
	panic("rm attempted to read Automation input")
}
