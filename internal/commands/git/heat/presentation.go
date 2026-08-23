package heat

import (
	"strings"
	"time"
)

const noChangedFilesMessage = "No changed files found in the selected range."

// RenderReport renders the report's semantic fields without owning terminal I/O.
func RenderReport(report Report, now time.Time, color bool) string {
	if report.IsEmpty() {
		return noChangedFilesMessage + "\n"
	}

	rows := report.Rows()
	var output strings.Builder
	output.WriteString(reportSummary(report, len(rows)))
	output.WriteByte('\n')
	output.WriteByte('\n')
	output.WriteString("#\tChanged at\tM A D R C\t")
	if report.Target == TargetFiles {
		output.WriteString("File\n")
	} else {
		output.WriteString("Directory\n")
	}

	marks := TimeMarks(rows)
	for index, row := range rows {
		output.WriteString(rankLabel(index+1, marks[index]))
		output.WriteByte('\t')
		output.WriteString(FormatChangedAt(row, report.RelativeTime, now))
		output.WriteByte('\t')
		output.WriteString(kindLabels(row))
		output.WriteByte('\t')
		output.WriteString(highlightPath(row.Path, QueryMatches(row.Path, report.Query), color))
		output.WriteByte('\n')
	}
	output.WriteString("Legend: latest, earliest, M modified, A added, D deleted, R renamed, C copied\n")
	return output.String()
}

func reportSummary(report Report, count int) string {
	parts := []string{report.RepositoryName, report.RangeLabel}
	if report.ShowCommitCount() {
		parts = append(parts, countLabel(report.CommitCount, "commit"))
	}
	if report.Target == TargetFiles {
		parts = append(parts, countLabel(count, "file"))
	} else {
		parts = append(parts, countLabel(count, "directory"))
	}
	return strings.Join(parts, " | ")
}

func countLabel(count int, singular string) string {
	if count == 1 {
		return integerString(count) + " " + singular
	}
	if singular == "directory" {
		return integerString(count) + " directories"
	}
	return integerString(count) + " " + singular + "s"
}

func rankLabel(rank int, mark TimeMark) string {
	if mark == "" {
		return integerString(rank)
	}
	return integerString(rank) + " (" + string(mark) + ")"
}

func kindLabels(row PathHeat) string {
	return strings.Join([]string{
		kindLabel(ChangeModified, row.Modified),
		kindLabel(ChangeAdded, row.Added),
		kindLabel(ChangeDeleted, row.Deleted),
		kindLabel(ChangeRenamed, row.Renamed),
		kindLabel(ChangeCopied, row.Copied),
	}, " ")
}

func kindLabel(kind ChangeKind, count int) string {
	if count == 0 {
		return "-"
	}
	return string(kind)
}

func highlightPath(filePath string, matches []Match, color bool) string {
	if !color || len(matches) == 0 {
		return filePath
	}
	var output strings.Builder
	end := 0
	for _, match := range matches {
		output.WriteString(filePath[end:match.Start])
		output.WriteString("\x1b[1;30;43m")
		output.WriteString(filePath[match.Start:match.End])
		output.WriteString("\x1b[0m")
		end = match.End
	}
	output.WriteString(filePath[end:])
	return output.String()
}
