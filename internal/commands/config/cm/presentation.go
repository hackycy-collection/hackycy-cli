package cm

import (
	"bytes"
	"fmt"
)

const emptyListMessage = "No CM profiles configured. Run \"ycy config cm add\" to add one."

// Render creates a stable noninteractive-friendly CM profile listing.
func Render(profiles []Profile) string {
	if len(profiles) == 0 {
		return emptyListMessage + "\n"
	}

	var output bytes.Buffer
	for _, profile := range profiles {
		marker := " "
		if profile.Default {
			marker = "*"
		}
		_, _ = fmt.Fprintf(&output, "%s %s %s %s\n", marker, profile.Name, profile.Model, profile.BaseURL)
	}
	return output.String()
}
