package main

import (
	"bytes"
	"errors"
	"testing"
	"time"

	heatcommand "github.com/hackycy/hackycy-cli/internal/commands/git/heat"
)

func TestTerminalGitHeatPresenterWritesTitleForReportsAndNotEmptyResults(t *testing.T) {
	output := &bytes.Buffer{}
	presenter := terminalGitHeatPresenter{output: output}
	report := heatcommand.Report{
		RepositoryName: "repo",
		RangeLabel:     "last 1 commits",
		Target:         heatcommand.TargetFiles,
		CommitCount:    1,
		Files: []heatcommand.PathHeat{{
			Path:      "file.txt",
			Counts:    heatcommand.Counts{Total: 1},
			ChangedAt: "2024-01-01 00:00:00",
		}},
	}
	if err := presenter.Present(report, time.Time{}); err != nil {
		t.Fatalf("Present() error = %v", err)
	}
	if got, want := output.String(), "HACKYCY CLI\n\nrepo | last 1 commits | 1 file\n\n#\tChanged at\tM A D R C\tFile\n1\t2024-01-01 00:00:00\t- - - - -\tfile.txt\nLegend: latest, earliest, M modified, A added, D deleted, R renamed, C copied\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}

	output.Reset()
	if err := presenter.Present(heatcommand.Report{Target: heatcommand.TargetFiles}, time.Time{}); err != nil {
		t.Fatalf("empty Present() error = %v", err)
	}
	if got, want := output.String(), "No changed files found in the selected range.\n"; got != want {
		t.Fatalf("empty output = %q, want %q", got, want)
	}
}

func TestTerminalGitHeatPresenterPropagatesWriterFailures(t *testing.T) {
	failure := errors.New("writer failed")
	presenter := terminalGitHeatPresenter{output: errorWriter{err: failure}}
	err := presenter.Present(heatcommand.Report{Target: heatcommand.TargetFiles}, time.Time{})
	if !errors.Is(err, failure) {
		t.Fatalf("Present() error = %v, want %v", err, failure)
	}
}

type errorWriter struct {
	err error
}

func (writer errorWriter) Write([]byte) (int, error) {
	return 0, writer.err
}
