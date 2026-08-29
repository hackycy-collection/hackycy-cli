package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

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
	app, err := newRootCommandForTest("0.0.0-dev", rootTestDependencies{
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

type panicRMReader struct{}

func (panicRMReader) Read([]byte) (int, error) {
	panic("rm attempted to read Automation input")
}
