package main

import (
	"bytes"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	pulsecommand "github.com/hackycy/hackycy-cli/internal/commands/git/pulse"
)

func TestTerminalPulsePrompterSelectsLegacyDayRangesAndAuthors(t *testing.T) {
	t.Run("day range", func(t *testing.T) {
		output := &bytes.Buffer{}
		prompter := newTerminalPulsePrompter(strings.NewReader("4\n"), output)
		days, cancelled := prompter.SelectDays(pulsecommand.DayPrompt{
			Message: "Select date range:",
			Options: []pulsecommand.DayChoice{
				{Value: 1, Label: "Today"},
				{Value: 2, Label: "Yesterday"},
				{Value: 3, Label: "Last 3 days"},
				{Value: 7, Label: "Last 7 days"},
			},
		})
		if cancelled || days != 7 {
			t.Fatalf("SelectDays() = (%d, %t), want 7, false", days, cancelled)
		}
		if got, want := output.String(), "Select date range:\n1) Today\n2) Yesterday\n3) Last 3 days\n4) Last 7 days\n> "; got != want {
			t.Fatalf("output = %q, want %q", got, want)
		}
	})

	t.Run("author defaults and required selection", func(t *testing.T) {
		output := &bytes.Buffer{}
		prompter := newTerminalPulsePrompter(strings.NewReader("\n2\n"), output)
		selected, cancelled := prompter.SelectAuthors(pulsecommand.AuthorPrompt{
			Message:  "Filter by authors:",
			Options:  []pulsecommand.AuthorChoice{{Value: "Ada", Label: "Ada"}, {Value: "Ben", Label: "Ben"}},
			Required: true,
		})
		if cancelled || !reflect.DeepEqual(selected, []string{"Ben"}) {
			t.Fatalf("SelectAuthors() = (%#v, %t), want Ben, false", selected, cancelled)
		}
		if got, want := output.String(), "Filter by authors:\n1) Ada\n2) Ben\n> At least one author is required.\n> "; got != want {
			t.Fatalf("output = %q, want %q", got, want)
		}

		output.Reset()
		prompter = newTerminalPulsePrompter(strings.NewReader("\n"), output)
		selected, cancelled = prompter.SelectAuthors(pulsecommand.AuthorPrompt{
			Options:       []pulsecommand.AuthorChoice{{Value: "Ada", Label: "Ada"}},
			InitialValues: []string{"Ada"},
			Required:      true,
		})
		if cancelled || !reflect.DeepEqual(selected, []string{"Ada"}) {
			t.Fatalf("default SelectAuthors() = (%#v, %t)", selected, cancelled)
		}
	})
}

func TestTerminalPulsePrompterTreatsCancellationAsSuccessfulPromptCancellation(t *testing.T) {
	prompter := newTerminalPulsePrompter(strings.NewReader("cancel\n"), &bytes.Buffer{})
	_, cancelled := prompter.SelectDays(pulsecommand.DayPrompt{Options: []pulsecommand.DayChoice{{Value: 1, Label: "Today"}}})
	if !cancelled {
		t.Fatal("SelectDays() did not cancel")
	}
}

func TestTerminalPulsePresenterRendersSemanticProgressAndReport(t *testing.T) {
	output := &bytes.Buffer{}
	presenter := terminalPulsePresenter{output: output}
	presenter.Introduction("/workspace")
	presenter.ScanStarted()
	presenter.RepositoryFound("/workspace", "/workspace/project", 1)
	presenter.RepositoriesFound(1)
	presenter.FetchStarted(1)
	presenter.FetchProgress("/workspace", "/workspace/project", 1, 1)
	presenter.Present(pulsecommand.Report{
		CommitCount: 1,
		Repositories: []pulsecommand.RepositoryReport{{
			Path:    "/workspace/project",
			Commits: []pulsecommand.Commit{{Date: "2026-08-23 10:00:00", Author: "Ada", Subject: "message"}},
		}},
	})
	want := "HACKYCY CLI\n\nGit Commit Tree\nWorkspace: /workspace\n" +
		"Scanning repositories...\nScanning repositories... [1] project\nFound 1 repository\n" +
		"Fetching commits... [0/1]\nFetching commits... [1/1] project\n\n" +
		"Found 1 commit in 1 repository\n\nproject (1 commit)\n   " + filepath.Dir("/workspace/project") + string(filepath.Separator) + "\n" +
		"   `- 2026-08-23 10:00:00 | Ada | message\n"
	if got := output.String(); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}
