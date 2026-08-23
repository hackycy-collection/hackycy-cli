package fork

import (
	"bytes"
	"fmt"
	"text/tabwriter"
)

const emptyListMessage = "No instances configured. Run \"ycy config fork add\" to add one."

// Render creates a stable noninteractive-friendly Fork instance listing.
func Render(instances []Instance) string {
	if len(instances) == 0 {
		return emptyListMessage + "\n"
	}

	var output bytes.Buffer
	table := tabwriter.NewWriter(&output, 0, 8, 2, ' ', 0)
	_, _ = fmt.Fprintln(table, "NAME\tTYPE\tSCHEME\tHOST\tTOKEN")
	for _, instance := range instances {
		_, _ = fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\n", instance.Name, instance.Type, instance.Scheme, instance.Host, instance.TokenPreview)
	}
	_ = table.Flush()

	label := "instances"
	if len(instances) == 1 {
		label = "instance"
	}
	_, _ = fmt.Fprintf(&output, "%d %s configured\n", len(instances), label)
	return output.String()
}
