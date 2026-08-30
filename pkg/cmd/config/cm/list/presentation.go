package list

import (
	"bytes"
	"fmt"

	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
)

const terminalCMListEmptyMessage = "No CM profiles configured. Run \"ycy config cm add\" to add one."

func terminalCMListDocument(result Result) terminalexperience.PresentationDocument {
	blocks := []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleTitle, Text: "Commit message profiles"}}
	if len(result.Profiles) == 0 {
		return terminalexperience.PresentationDocument{Blocks: append(blocks,
			terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleWarning, Text: "No CM profiles configured."},
			terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleMuted, Text: "Run \"ycy config cm add\" to add one."},
		)}
	}
	blocks = append(blocks, terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleMuted, Text: "PROFILE  MODEL  BASE URL"})
	for _, profile := range result.Profiles {
		role := terminalexperience.VisualRolePlain
		if profile.Default {
			role = terminalexperience.VisualRoleSuccess
		}
		blocks = append(blocks, terminalexperience.PresentationBlock{Role: role, Text: terminalCMListRow(profile)})
	}
	return terminalexperience.PresentationDocument{Blocks: blocks}
}

func terminalCMListPlainText(profiles []Profile) string {
	if len(profiles) == 0 {
		return terminalCMListEmptyMessage + "\n"
	}
	var output bytes.Buffer
	for _, profile := range profiles {
		_, _ = fmt.Fprintln(&output, terminalCMListRow(profile))
	}
	return output.String()
}

func terminalCMListRow(profile Profile) string {
	marker := " "
	if profile.Default {
		marker = "*"
	}
	return fmt.Sprintf("%s %s %s %s", marker, profile.Name, profile.Model, profile.BaseURL)
}
