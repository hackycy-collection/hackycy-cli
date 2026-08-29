package pulse

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/hackycy/hackycy-cli/internal/gitprocess"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestTerminalPulseAdapterTranslatesFormsPhasesAndPresentation(t *testing.T) {
	experience := terminaltest.NewRecordingExperience(
		terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Value: "7"}},
		terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Values: []string{"Ada"}}},
	)
	run := experience.Open(context.Background())
	adapter := newTerminalPulseAdapter(run, terminalexperience.Session{Kind: terminalexperience.RichInteractive, Color: true}, func() {})

	days, cancelled, err := adapter.SelectDays(DayPrompt{Message: "Select date range:", Options: []DayChoice{{Value: 1, Label: "Today"}, {Value: 7, Label: "Last 7 days"}}})
	if err != nil || cancelled || days != 7 {
		t.Fatalf("SelectDays() = (%d, %t, %v)", days, cancelled, err)
	}
	authors, cancelled, err := adapter.SelectAuthors(AuthorPrompt{Message: "Filter by authors:", Options: []AuthorChoice{{Value: "Ada", Label: "Ada"}, {Value: "Ben", Label: "Ben"}}, InitialValues: []string{"Ada", "Ben"}, Required: true})
	if err != nil || cancelled || !reflect.DeepEqual(authors, []string{"Ada"}) {
		t.Fatalf("SelectAuthors() = (%#v, %t, %v)", authors, cancelled, err)
	}
	adapter.Introduction("/workspace")
	adapter.RepositoriesFound(2)
	adapter.Present(Report{CommitCount: 1, Repositories: []RepositoryReport{{Path: "/workspace/project", Commits: []Commit{{Date: "2026-08-23 10:00:00", Author: "Ada", Subject: "message"}}}}})
	if err := run.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	operations := experience.Run.Operations()
	if len(operations) != 6 || operations[0].Kind != terminaltest.AskOperation || operations[1].Kind != terminaltest.AskOperation || operations[5].Kind != terminaltest.CloseOperation {
		t.Fatalf("operations = %#v", operations)
	}
	daysRequest := operations[0].Value.(terminalexperience.InteractionRequest)
	if daysRequest.Kind != terminalexperience.InteractionSelect || daysRequest.Message != "Select date range:" || !reflect.DeepEqual(daysRequest.CancelValues, []string{"", "q", "quit", "cancel"}) {
		t.Fatalf("days request = %#v", daysRequest)
	}
	authorRequest := operations[1].Value.(terminalexperience.InteractionRequest)
	if authorRequest.Kind != terminalexperience.InteractionMultiSelect || !authorRequest.HasDefault || !reflect.DeepEqual(authorRequest.Default.Values, []string{"Ada", "Ben"}) {
		t.Fatalf("author request = %#v", authorRequest)
	}
	intro := operations[2].Value.(terminalexperience.PresentationDocument)
	if !reflect.DeepEqual(intro.Blocks, []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleTitle, Text: "HACKYCY CLI"}, {Role: terminalexperience.VisualRoleActive, Text: "Git Commit Tree"}, {Role: terminalexperience.VisualRoleMuted, Text: "Workspace: /workspace"}}) {
		t.Fatalf("intro = %#v", intro)
	}
	if report := operations[4].Value.(terminalexperience.PresentationDocument); report.Blocks[0].Role != terminalexperience.VisualRoleSuccess || report.Blocks[0].Text != "Found 1 commit in 1 repository" {
		t.Fatalf("report = %#v", report)
	}
}

func TestTerminalPulseAdapterMapsCancellationAndAutomationInteraction(t *testing.T) {
	cancelledExperience := terminaltest.NewRecordingExperience(terminaltest.SemanticAnswer{Err: terminalexperience.ErrInteractionCancelled})
	adapter := newTerminalPulseAdapter(cancelledExperience.Open(context.Background()), terminalexperience.Session{Kind: terminalexperience.RichInteractive}, func() {})
	if _, cancelled, err := adapter.SelectDays(DayPrompt{}); err != nil || !cancelled {
		t.Fatalf("cancelled SelectDays() = (%t, %v)", cancelled, err)
	}

	automationExperience := terminaltest.NewRecordingExperience(terminaltest.SemanticAnswer{Err: terminalexperience.ErrAutomationInteraction})
	automation := newTerminalPulseAdapter(automationExperience.Open(context.Background()), terminalexperience.Session{Kind: terminalexperience.Automation}, func() {})
	if _, _, err := automation.SelectDays(DayPrompt{}); !errors.Is(err, errGitPulseRequiresInteractive) {
		t.Fatalf("Automation SelectDays() error = %v", err)
	}
}

func TestTerminalPulseAdapterBridgesTypedPhaseToTrackedOperation(t *testing.T) {
	experience := terminaltest.NewRecordingExperience()
	adapter := newTerminalPulseAdapter(experience.Open(context.Background()), terminalexperience.Session{Kind: terminalexperience.RichInteractive}, func() {})
	reporter, err := adapter.Start(context.Background(), PhaseScan)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	reporter.Report(Phase{Kind: PhaseScan, State: PhaseActive, Root: "/workspace", Repository: "/workspace/project", Completed: 1})
	if err := reporter.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	operations := experience.Run.Operations()
	if len(operations) != 1 || operations[0].Kind != terminaltest.TrackOperation {
		t.Fatalf("operations = %#v", operations)
	}
	operation := operations[0].Value.(terminalexperience.TrackedOperation)
	if operation.Label != "Git Pulse" {
		t.Fatalf("tracked operation = %#v", operation)
	}
}

func TestRunPulseAutomationFailsBeforeReadingInputOrWritingAResult(t *testing.T) {
	workspace := t.TempDir()
	initializeStandalonePulseRepository(t, filepath.Join(workspace, "alpha"), "Ada", "ada@example.test", "alpha commit")
	initializeStandalonePulseRepository(t, filepath.Join(workspace, "beta"), "Ben", "ben@example.test", "beta commit")

	for _, testCase := range []struct {
		name string
		days *int
	}{
		{name: "date selection"},
		{name: "author selection", days: pulseInt(1)},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
				Session:     terminalexperience.Session{Kind: terminalexperience.Automation},
				Input:       panicPulseReader{},
				Output:      stdout,
				Diagnostics: stderr,
			})
			err := runPulse(&Options{
				Context:          context.Background(),
				Directory:        workspace,
				Days:             testCase.days,
				WorkingDirectory: func() (string, error) { return workspace, nil },
				Terminal:         experience,
				Git:              &gitprocess.Runner{},
				Now:              time.Now,
			})
			if !errors.Is(err, errGitPulseRequiresInteractive) || stdout.Len() != 0 || stderr.Len() != 0 {
				t.Fatalf("Run() error = %v, streams = (%q, %q)", err, stdout.String(), stderr.String())
			}
			if terminaltest.ContainsTerminalControl(append(stdout.Bytes(), stderr.Bytes()...)) {
				t.Fatalf("Automation streams contain terminal control: (%q, %q)", stdout.String(), stderr.String())
			}
		})
	}
}

func TestRunPulsePlainJourneyKeepsPhasesOnStderrAndReportOnStdout(t *testing.T) {
	workspace := t.TempDir()
	initializeStandalonePulseRepository(t, filepath.Join(workspace, "alpha"), "Ada", "ada@example.test", "alpha commit")
	initializeStandalonePulseRepository(t, filepath.Join(workspace, "beta"), "Ben", "ben@example.test", "beta commit")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Session:     terminalexperience.Session{Kind: terminalexperience.PlainInteractive},
		Input:       strings.NewReader("1\n1,2\n"),
		Output:      stdout,
		Diagnostics: stderr,
	})

	err := runPulse(&Options{
		Context:          context.Background(),
		Directory:        workspace,
		WorkingDirectory: func() (string, error) { return workspace, nil },
		Terminal:         experience,
		Git:              &gitprocess.Runner{},
		Now:              time.Now,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	for _, expected := range []string{"HACKYCY CLI", "Git Commit Tree", "Found 2 repositories", "Found 2 commits in 2 repositories", "alpha commit", "beta commit"} {
		if !strings.Contains(stdout.String(), expected) {
			t.Fatalf("stdout missing %q: %q", expected, stdout.String())
		}
	}
	for _, expected := range []string{"Scanning repositories", "Select date range:", "Fetching commits", "Filter by authors:"} {
		if !strings.Contains(stderr.String(), expected) {
			t.Fatalf("stderr missing %q: %q", expected, stderr.String())
		}
	}
	if terminaltest.ContainsTerminalControl(append(stdout.Bytes(), stderr.Bytes()...)) {
		t.Fatalf("Plain streams contain terminal control: (%q, %q)", stdout.String(), stderr.String())
	}
}

type panicPulseReader struct{}

func (panicPulseReader) Read([]byte) (int, error) {
	panic("git pulse Automation must not read stdin")
}
