package heat

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
)

const (
	terminalGitHeatEmptyMessage = "No changed files found in the selected range."
	terminalGitHeatLegend       = "Legend: latest, earliest, M modified, A added, D deleted, R renamed, C copied"
)

func runHeat(options *Options) error {
	if options == nil || options.Context == nil || options.Terminal == nil || options.Git == nil || options.Now == nil {
		return errors.New("git heat options are incomplete")
	}
	module, err := New(Dependencies{
		Git: gitRunnerAdapter{runner: options.Git},
		Now: options.Now,
	})
	if err != nil {
		return err
	}
	result, err := module.Run(options.Context, Input{
		Limit:        options.Limit,
		Days:         options.Days,
		Target:       options.Target,
		Sort:         options.Sort,
		RelativeTime: options.RelativeTime,
		Query:        options.Query,
	})
	if err != nil {
		return err
	}
	run := options.Terminal.Open(options.Context)
	defer run.Close()
	return run.Present(terminalGitHeatDocument(options.Terminal.Session(), result))
}

func terminalGitHeatDocument(session terminalexperience.Session, result Result) terminalexperience.PresentationDocument {
	report := result.Report
	if session.Kind != terminalexperience.RichInteractive {
		return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{
			Role: terminalexperience.VisualRolePlain,
			Text: terminalGitHeatPlainText(report, result.Now),
		}}}
	}
	if report.IsEmpty() {
		return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{
			Role: terminalexperience.VisualRoleWarning,
			Text: terminalGitHeatEmptyMessage,
		}}}
	}

	rows := report.Rows()
	marks := TimeMarks(rows)
	blocks := []terminalexperience.PresentationBlock{{
		Role: terminalexperience.VisualRoleTitle,
		Text: "HACKYCY CLI",
	}, {
		Role: terminalexperience.VisualRoleActive,
		Text: terminalGitHeatSummary(report, len(rows)),
	}, {
		Role: terminalexperience.VisualRoleMuted,
		Text: terminalGitHeatHeader(report),
	}}
	for index, row := range rows {
		role := terminalexperience.VisualRolePlain
		if len(QueryMatches(row.Path, report.Query)) > 0 {
			role = terminalexperience.VisualRoleActive
		}
		blocks = append(blocks, terminalexperience.PresentationBlock{
			Role: role,
			Text: terminalGitHeatRow(index+1, marks[index], row, report.RelativeTime, result.Now),
		})
	}
	blocks = append(blocks, terminalexperience.PresentationBlock{
		Role: terminalexperience.VisualRoleMuted,
		Text: terminalGitHeatLegend,
	})
	return terminalexperience.PresentationDocument{ClearBefore: true, Blocks: blocks}
}

func terminalGitHeatPlainText(report Report, now time.Time) string {
	if report.IsEmpty() {
		return terminalGitHeatEmptyMessage + "\n"
	}
	return "HACKYCY CLI\n\n" + terminalGitHeatReportText(report, now)
}

func terminalGitHeatReportText(report Report, now time.Time) string {
	rows := report.Rows()
	marks := TimeMarks(rows)
	var output strings.Builder
	output.WriteString(terminalGitHeatSummary(report, len(rows)))
	output.WriteString("\n\n")
	output.WriteString(terminalGitHeatHeader(report))
	output.WriteByte('\n')
	for index, row := range rows {
		output.WriteString(terminalGitHeatRow(index+1, marks[index], row, report.RelativeTime, now))
		output.WriteByte('\n')
	}
	output.WriteString(terminalGitHeatLegend)
	output.WriteByte('\n')
	return output.String()
}

func terminalGitHeatSummary(report Report, count int) string {
	parts := []string{report.RepositoryName, report.RangeLabel}
	if report.ShowCommitCount() {
		parts = append(parts, terminalGitHeatCountLabel(report.CommitCount, "commit"))
	}
	if report.Target == TargetFiles {
		parts = append(parts, terminalGitHeatCountLabel(count, "file"))
	} else {
		parts = append(parts, terminalGitHeatCountLabel(count, "directory"))
	}
	return strings.Join(parts, " | ")
}

func terminalGitHeatCountLabel(count int, singular string) string {
	if count == 1 {
		return strconv.Itoa(count) + " " + singular
	}
	if singular == "directory" {
		return strconv.Itoa(count) + " directories"
	}
	return strconv.Itoa(count) + " " + singular + "s"
}

func terminalGitHeatHeader(report Report) string {
	if report.Target == TargetFiles {
		return "#\tChanged at\tM A D R C\tFile"
	}
	return "#\tChanged at\tM A D R C\tDirectory"
}

func terminalGitHeatRow(rank int, mark TimeMark, row PathHeat, relative bool, now time.Time) string {
	return fmt.Sprintf("%s\t%s\t%s\t%s",
		terminalGitHeatRankLabel(rank, mark),
		FormatChangedAt(row, relative, now),
		terminalGitHeatKindLabels(row),
		row.Path,
	)
}

func terminalGitHeatRankLabel(rank int, mark TimeMark) string {
	if mark == "" {
		return strconv.Itoa(rank)
	}
	return strconv.Itoa(rank) + " (" + string(mark) + ")"
}

func terminalGitHeatKindLabels(row PathHeat) string {
	return strings.Join([]string{
		terminalGitHeatKindLabel("M", row.Modified),
		terminalGitHeatKindLabel("A", row.Added),
		terminalGitHeatKindLabel("D", row.Deleted),
		terminalGitHeatKindLabel("R", row.Renamed),
		terminalGitHeatKindLabel("C", row.Copied),
	}, " ")
}

func terminalGitHeatKindLabel(label string, count int) string {
	if count == 0 {
		return "-"
	}
	return label
}
