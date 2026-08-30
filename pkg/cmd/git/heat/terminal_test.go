package heat

import (
	"bytes"
	"context"
	"testing"
	"time"

	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestTerminalGitHeatPresentationPreservesPlainAndAutomationResults(t *testing.T) {
	report := terminalGitHeatTestReport()
	for _, testCase := range []struct {
		name   string
		result Result
		want   string
	}{
		{
			name:   "report",
			result: Result{Report: report},
			want:   "HACKYCY CLI\nrepo | last 1 commits | 1 file\n#\tChanged at\tM A D R C\tFile\n1\t2024-01-01 00:00:00\t- - - - -\tfile.txt\nLegend: latest, earliest, M modified, A added, D deleted, R renamed, C copied\n",
		},
		{
			name:   "empty",
			result: Result{Report: Report{Target: TargetFiles}},
			want:   "No changed files found in the selected range.\n",
		},
	} {
		for _, session := range []terminalexperience.Capabilities{
			{Interaction: terminalexperience.PlainInteractive},
			{Interaction: terminalexperience.Automation},
		} {
			var output bytes.Buffer
			experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{Capabilities: session, Output: &output})
			run := experience.Open(context.Background())
			if err := run.Result(terminalGitHeatDocument(testCase.result)); err != nil {
				t.Fatalf("%s Present() error = %v", testCase.name, err)
			}
			if err := run.Close(); err != nil {
				t.Fatalf("%s Close() error = %v", testCase.name, err)
			}
			if got := output.String(); got != testCase.want {
				t.Fatalf("%s %v output = %q, want %q", testCase.name, session.Interaction, got, testCase.want)
			}
			if terminaltest.ContainsTerminalControl(output.Bytes()) {
				t.Fatalf("%s %v output contains terminal control: %q", testCase.name, session.Interaction, output.String())
			}
		}
	}
}

func TestTerminalGitHeatPresentationUsesRichSemanticRoles(t *testing.T) {
	result := Result{Report: terminalGitHeatTestReport(), Now: time.Time{}}
	document := terminalGitHeatDocument(result)
	want := []terminalexperience.VisualRole{
		terminalexperience.VisualRoleTitle,
		terminalexperience.VisualRoleActive,
		terminalexperience.VisualRoleMuted,
		terminalexperience.VisualRoleActive,
		terminalexperience.VisualRoleMuted,
	}
	if len(document.Blocks) != len(want) {
		t.Fatalf("blocks = %#v", document.Blocks)
	}
	for index, role := range want {
		if document.Blocks[index].Role != role {
			t.Fatalf("block %d role = %v, want %v", index, document.Blocks[index].Role, role)
		}
	}
}

func terminalGitHeatTestReport() Report {
	return Report{
		RepositoryName: "repo",
		RangeLabel:     "last 1 commits",
		Target:         TargetFiles,
		Query:          "file",
		CommitCount:    1,
		Files: []PathHeat{{
			Path:      "file.txt",
			Counts:    Counts{Total: 1},
			ChangedAt: "2024-01-01 00:00:00",
		}},
	}
}
