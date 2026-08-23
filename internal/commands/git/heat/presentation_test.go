package heat

import (
	"strings"
	"testing"
	"time"
)

func TestRenderReportIncludesAllSemanticFields(t *testing.T) {
	report := Report{
		RepositoryName: "repo",
		RangeLabel:     "last 2 days",
		Target:         TargetDirectories,
		RelativeTime:   true,
		Query:          "src",
		CommitCount:    2,
		Files: []PathHeat{
			{Path: "src/main.go", Counts: Counts{Total: 2, Modified: 1, Added: 1}, ChangedAt: "2024-03-10 00:00:00", ChangedAtEpoch: 100},
			{Path: "README.md", Counts: Counts{Total: 1, Deleted: 1}, ChangedAt: "2024-03-09 00:00:00", ChangedAtEpoch: 50},
		},
		Directories: []PathHeat{
			{Path: "src", Counts: Counts{Total: 2, Modified: 1, Added: 1}, ChangedAt: "2024-03-10 00:00:00", ChangedAtEpoch: 100},
			{Path: ".", Counts: Counts{Total: 1, Deleted: 1}, ChangedAt: "2024-03-09 00:00:00", ChangedAtEpoch: 50},
		},
	}
	output := RenderReport(report, time.Unix(160, 0), false)
	for _, expected := range []string{
		"repo | last 2 days | 2 commits | 2 directories",
		"#\tChanged at\tM A D R C\tDirectory",
		"1 (latest)\t1 minute ago\tM A - - -\tsrc",
		"2 (earliest)\t1 minute ago\t- - D - -\t.",
		"Legend: latest, earliest, M modified, A added, D deleted, R renamed, C copied",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("output does not contain %q:\n%s", expected, output)
		}
	}
	if strings.Contains(output, "\x1b[") {
		t.Fatalf("NO_COLOR output contains ANSI: %q", output)
	}
}

func TestRenderReportUsesPathHighlightOnlyWhenColorIsEnabled(t *testing.T) {
	report := Report{
		RepositoryName: "repo",
		RangeLabel:     "last 1 commits",
		Target:         TargetFiles,
		Query:          "API",
		CommitCount:    1,
		Files:          []PathHeat{{Path: "src/api.go", Counts: Counts{Total: 1}, ChangedAt: "now"}},
	}
	plain := RenderReport(report, time.Time{}, false)
	if !strings.Contains(plain, "src/api.go") || strings.Contains(plain, "\x1b[") {
		t.Fatalf("plain output = %q", plain)
	}
	colored := RenderReport(report, time.Time{}, true)
	if !strings.Contains(colored, "src/\x1b[1;30;43mapi\x1b[0m.go") {
		t.Fatalf("colored output = %q", colored)
	}
}

func TestRenderReportMapsEmptyResultsToTheInformationalMessage(t *testing.T) {
	output := RenderReport(Report{Target: TargetFiles}, time.Time{}, false)
	if output != noChangedFilesMessage+"\n" {
		t.Fatalf("empty output = %q", output)
	}
}
