package list

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"

	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
)

const terminalForkListEmptyMessage = "No instances configured. Run \"ycy config fork add\" to add one."

func terminalForkListDocument(session terminalexperience.Session, result Result) terminalexperience.PresentationDocument {
	if session.Kind != terminalexperience.RichInteractive {
		return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{
			Role: terminalexperience.VisualRolePlain,
			Text: terminalForkListPlainText(result.Instances),
		}}}
	}

	blocks := []terminalexperience.PresentationBlock{{
		Role: terminalexperience.VisualRoleTitle,
		Text: "Fork provider instances",
	}}
	if len(result.Instances) == 0 {
		return terminalexperience.PresentationDocument{Blocks: append(blocks,
			terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleWarning, Text: "No instances configured."},
			terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleMuted, Text: "Run \"ycy config fork add\" to add one."},
		)}
	}
	blocks = append(blocks,
		terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleMuted, Text: "NAME  TYPE  SCHEME  HOST  TOKEN"},
		terminalexperience.PresentationBlock{Role: terminalexperience.VisualRolePlain, Text: terminalForkListRows(result.Instances)},
		terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleSuccess, Text: terminalForkListCount(len(result.Instances))},
	)
	return terminalexperience.PresentationDocument{Blocks: blocks}
}

func terminalForkListPlainText(instances []Instance) string {
	if len(instances) == 0 {
		return terminalForkListEmptyMessage + "\n"
	}

	var output bytes.Buffer
	table := tabwriter.NewWriter(&output, 0, 8, 2, ' ', 0)
	_, _ = fmt.Fprintln(table, "NAME\tTYPE\tSCHEME\tHOST\tTOKEN")
	for _, instance := range instances {
		_, _ = fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\n", instance.Name, instance.Type, instance.Scheme, instance.Host, instance.TokenPreview)
	}
	_ = table.Flush()
	_, _ = fmt.Fprintln(&output, terminalForkListCount(len(instances)))
	return output.String()
}

func terminalForkListRows(instances []Instance) string {
	lines := make([]string, 0, len(instances))
	for _, instance := range instances {
		lines = append(lines, fmt.Sprintf("%s  %s  %s  %s  %s", instance.Name, instance.Type, instance.Scheme, instance.Host, instance.TokenPreview))
	}
	return strings.Join(lines, "\n")
}

func terminalForkListCount(count int) string {
	label := "instances"
	if count == 1 {
		label = "instance"
	}
	return fmt.Sprintf("%d %s configured", count, label)
}
