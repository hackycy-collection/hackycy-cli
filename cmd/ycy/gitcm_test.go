package main

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	cmcommand "github.com/hackycy/hackycy-cli/internal/commands/git/cm"
)

func TestTerminalGitCMPrompterSelectsDefaultsIndicesAndCancellation(t *testing.T) {
	prompt := cmcommand.StagePrompt{
		Message:       "Select files to stage",
		Options:       []cmcommand.StageOption{{Value: "one.go", Label: "M one.go"}, {Value: "two.go", Label: "A two.go"}},
		InitialValues: []string{"one.go", "two.go"},
	}
	prompter := newTerminalGitCMPrompter(strings.NewReader("2\n"), &bytes.Buffer{})
	selected, cancelled := prompter.SelectFiles(prompt)
	if cancelled || !reflect.DeepEqual(selected, []string{"two.go"}) {
		t.Fatalf("SelectFiles() = (%#v, %t)", selected, cancelled)
	}
	defaults := newTerminalGitCMPrompter(strings.NewReader("\n"), &bytes.Buffer{})
	selected, cancelled = defaults.SelectFiles(prompt)
	if cancelled || !reflect.DeepEqual(selected, prompt.InitialValues) {
		t.Fatalf("default SelectFiles() = (%#v, %t)", selected, cancelled)
	}
	none := newTerminalGitCMPrompter(strings.NewReader("none\n"), &bytes.Buffer{})
	selected, cancelled = none.SelectFiles(prompt)
	if cancelled || len(selected) != 0 {
		t.Fatalf("none SelectFiles() = (%#v, %t)", selected, cancelled)
	}
	cancel := newTerminalGitCMPrompter(strings.NewReader("cancel\n"), &bytes.Buffer{})
	_, cancelled = cancel.SelectFiles(prompt)
	if !cancelled {
		t.Fatal("SelectFiles() did not cancel")
	}
}

func TestTerminalGitCMPrompterConfirmsByDefaultAndShowsGeneratedMessage(t *testing.T) {
	output := &bytes.Buffer{}
	prompter := newTerminalGitCMPrompter(strings.NewReader("\n"), output)
	prompt := cmcommand.CommitPrompt{
		Message: "Create this commit?",
		Generated: cmcommand.GeneratedMessage{
			Message:  "feat(cm): present output",
			Usage:    gitCMUsage(1234, 42, 1276),
			Evidence: cmcommand.EvidenceCoverage{EstimatedLocalPromptTokens: 456, RepresentedClusters: 1, TotalClusters: 1, IncludedFacts: 4},
		},
		Profile: cmcommand.ProfileDiagnostic{Name: "work", Model: "model"},
	}
	confirmed, cancelled := prompter.ConfirmCommit(prompt)
	if !confirmed || cancelled {
		t.Fatalf("ConfirmCommit() = (%t, %t)", confirmed, cancelled)
	}
	if got := output.String(); !strings.Contains(got, "feat(cm): present output") || !strings.Contains(got, "Provider tokens: 1,234 prompt / 42 completion / 1,276 total") || !strings.Contains(got, "Create this commit? [Y/n]:") {
		t.Fatalf("prompt output = %q", got)
	}
	decline := newTerminalGitCMPrompter(strings.NewReader("no\n"), &bytes.Buffer{})
	confirmed, cancelled = decline.ConfirmCommit(prompt)
	if confirmed || cancelled {
		t.Fatalf("declined confirmation = (%t, %t)", confirmed, cancelled)
	}
	eof := newTerminalGitCMPrompter(strings.NewReader(""), &bytes.Buffer{})
	confirmed, cancelled = eof.ConfirmCommit(prompt)
	if confirmed || !cancelled {
		t.Fatalf("EOF confirmation = (%t, %t)", confirmed, cancelled)
	}
}

func TestTerminalGitCMPresenterMapsNormalResults(t *testing.T) {
	for _, testCase := range []struct {
		result cmcommand.Result
		want   string
	}{
		{result: cmcommand.Result{NoChanges: true, NoChangeScope: cmcommand.ScopeAllUncommitted}, want: "No uncommitted changes.\n"},
		{result: cmcommand.Result{NoChanges: true, NoChangeScope: cmcommand.ScopeStaged}, want: "No staged changes.\n"},
		{result: cmcommand.Result{NothingSelected: true}, want: "Nothing selected.\n"},
		{result: cmcommand.Result{Cancelled: true}, want: "Cancelled\n"},
		{result: cmcommand.Result{Committed: true}, want: "Commit created\n"},
		{result: cmcommand.Result{Pushed: true}, want: "Commit created and pushed\n"},
	} {
		output := &bytes.Buffer{}
		terminalGitCMPresenter{output: output}.Outcome(testCase.result)
		if output.String() != testCase.want {
			t.Fatalf("Outcome(%#v) = %q, want %q", testCase.result, output.String(), testCase.want)
		}
	}
}

func TestTerminalGitCMPresenterFormatsGeneratedCoverage(t *testing.T) {
	output := &bytes.Buffer{}
	presenter := terminalGitCMPresenter{output: output}
	presenter.Generated(cmcommand.Result{
		Generated: &cmcommand.GeneratedMessage{Message: "feat(cm): compact", Evidence: cmcommand.EvidenceCoverage{EstimatedLocalPromptTokens: 4000, RepresentedClusters: 2, TotalClusters: 3, IncludedFacts: 18, OmittedFacts: 13, ContentCompacted: true}},
		Profile:   cmcommand.ProfileDiagnostic{Name: "work", Model: "model"},
	})
	if got := output.String(); !strings.Contains(got, "Provider tokens: unavailable") || !strings.Contains(got, "4,000") || !strings.Contains(got, "3 clusters represented with compacted semantic evidence") {
		t.Fatalf("Generated() output = %q", got)
	}
}

func gitCMUsage(prompt, completion, total float64) *cmcommand.TokenUsage {
	return &cmcommand.TokenUsage{PromptTokens: &prompt, CompletionTokens: &completion, TotalTokens: &total}
}
