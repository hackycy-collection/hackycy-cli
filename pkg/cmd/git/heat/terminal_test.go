package heat

import (
	"bytes"
	"context"
	"strings"
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
			want:   "HACKYCY CLI\n\nrepo | last 1 commits | 1 file\n\n#\tChanged at\tM A D R C\tFile\n1\t2024-01-01 00:00:00\t- - - - -\tfile.txt\nLegend: latest, earliest, M modified, A added, D deleted, R renamed, C copied\n",
		},
		{
			name:   "empty",
			result: Result{Report: Report{Target: TargetFiles}},
			want:   "No changed files found in the selected range.\n",
		},
	} {
		for _, session := range []terminalexperience.Session{
			{Kind: terminalexperience.PlainInteractive},
			{Kind: terminalexperience.Automation},
		} {
			var output bytes.Buffer
			experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{Session: session, Output: &output})
			run := experience.Open(context.Background())
			if err := run.Present(terminalGitHeatDocument(session, testCase.result)); err != nil {
				t.Fatalf("%s Present() error = %v", testCase.name, err)
			}
			if err := run.Close(); err != nil {
				t.Fatalf("%s Close() error = %v", testCase.name, err)
			}
			if got := output.String(); got != testCase.want {
				t.Fatalf("%s %v output = %q, want %q", testCase.name, session.Kind, got, testCase.want)
			}
			if terminaltest.ContainsTerminalControl(output.Bytes()) {
				t.Fatalf("%s %v output contains terminal control: %q", testCase.name, session.Kind, output.String())
			}
		}
	}
}

func TestTerminalGitHeatPresentationUsesRichSemanticRoles(t *testing.T) {
	result := Result{Report: terminalGitHeatTestReport(), Now: time.Time{}}
	for _, session := range []terminalexperience.Session{
		{Kind: terminalexperience.RichInteractive, Color: true},
		{Kind: terminalexperience.RichInteractive},
	} {
		document := terminalGitHeatDocument(session, result)
		if !document.ClearBefore {
			t.Fatal("Rich report did not retain its title clear")
		}
		want := []terminalexperience.VisualRole{
			terminalexperience.VisualRoleTitle,
			terminalexperience.VisualRoleActive,
			terminalexperience.VisualRoleMuted,
			terminalexperience.VisualRoleActive,
			terminalexperience.VisualRoleMuted,
		}
		if len(document.Blocks) != len(want) {
			t.Fatalf("Rich blocks = %#v", document.Blocks)
		}
		for index, role := range want {
			if document.Blocks[index].Role != role {
				t.Fatalf("Rich block %d role = %v, want %v", index, document.Blocks[index].Role, role)
			}
		}
		var output bytes.Buffer
		experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{Session: session, Output: &output})
		run := experience.Open(context.Background())
		if err := run.Present(document); err != nil {
			t.Fatalf("Present() error = %v", err)
		}
		if err := run.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
		const clear = "\x1b[2J\x1b[H"
		if !strings.HasPrefix(output.String(), clear) {
			t.Fatalf("Rich output omitted title clear: %q", output.String())
		}
		if !session.Color && strings.Contains(output.String()[len(clear):], "\x1b[") {
			t.Fatalf("NO_COLOR Rich output contains style bytes: %q", output.String())
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
