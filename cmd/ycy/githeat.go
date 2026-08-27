package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hackycy/hackycy-cli/internal/cliapp"
	heatcommand "github.com/hackycy/hackycy-cli/internal/commands/git/heat"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
)

const (
	terminalGitHeatEmptyMessage = "No changed files found in the selected range."
	terminalGitHeatLegend       = "Legend: latest, earliest, M modified, A added, D deleted, R renamed, C copied"
)

func newGitHeatHandler(experience *terminalexperience.Runtime) (cliapp.GitHeatHandler, error) {
	module, err := heatcommand.New(heatcommand.Dependencies{
		Git: newOSHeatGitRunner(),
		Now: time.Now,
	})
	if err != nil {
		return nil, err
	}
	return func(ctx context.Context, input heatcommand.Input) (heatcommand.Result, error) {
		result, err := module.Run(ctx, input)
		if err != nil {
			return heatcommand.Result{}, err
		}
		run := experience.Open(ctx)
		defer run.Close()
		if err := run.Present(terminalGitHeatDocument(experience.Session(), result)); err != nil {
			return heatcommand.Result{}, err
		}
		return result, nil
	}, nil
}

func terminalGitHeatDocument(session terminalexperience.Session, result heatcommand.Result) terminalexperience.PresentationDocument {
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
	marks := heatcommand.TimeMarks(rows)
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
		if len(heatcommand.QueryMatches(row.Path, report.Query)) > 0 {
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

func terminalGitHeatPlainText(report heatcommand.Report, now time.Time) string {
	if report.IsEmpty() {
		return terminalGitHeatEmptyMessage + "\n"
	}
	return "HACKYCY CLI\n\n" + terminalGitHeatReportText(report, now)
}

func terminalGitHeatReportText(report heatcommand.Report, now time.Time) string {
	rows := report.Rows()
	marks := heatcommand.TimeMarks(rows)
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

func terminalGitHeatSummary(report heatcommand.Report, count int) string {
	parts := []string{report.RepositoryName, report.RangeLabel}
	if report.ShowCommitCount() {
		parts = append(parts, terminalGitHeatCountLabel(report.CommitCount, "commit"))
	}
	if report.Target == heatcommand.TargetFiles {
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

func terminalGitHeatHeader(report heatcommand.Report) string {
	if report.Target == heatcommand.TargetFiles {
		return "#\tChanged at\tM A D R C\tFile"
	}
	return "#\tChanged at\tM A D R C\tDirectory"
}

func terminalGitHeatRow(rank int, mark heatcommand.TimeMark, row heatcommand.PathHeat, relative bool, now time.Time) string {
	return fmt.Sprintf("%s\t%s\t%s\t%s",
		terminalGitHeatRankLabel(rank, mark),
		heatcommand.FormatChangedAt(row, relative, now),
		terminalGitHeatKindLabels(row),
		row.Path,
	)
}

func terminalGitHeatRankLabel(rank int, mark heatcommand.TimeMark) string {
	if mark == "" {
		return strconv.Itoa(rank)
	}
	return strconv.Itoa(rank) + " (" + string(mark) + ")"
}

func terminalGitHeatKindLabels(row heatcommand.PathHeat) string {
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
