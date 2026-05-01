package overlaydiff

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/tabwriter"
)

// FormatTable renders changes as a tabulated text report.
func FormatTable(changes []Change) string {
	if len(changes) == 0 {
		return "No differences found\n"
	}

	var out bytes.Buffer
	tw := tabwriter.NewWriter(&out, 0, 0, 2, ' ', 0)

	_, _ = fmt.Fprintln(tw, "CHANGE\tTYPE\tMODE\tSIZE\tPATH")
	for _, change := range changes {
		size := "-"
		if change.Type == EntryFile {
			size = fmt.Sprintf("%d", change.Size)
		}

		_, _ = fmt.Fprintf(tw, "%s\t%s\t%#o\t%s\t%s\n", change.Change, change.Type, change.Mode, size, change.Path)
	}

	_ = tw.Flush()
	return out.String()
}

// FormatJSON renders changes as indented JSON.
func FormatJSON(changes []Change) ([]byte, error) {
	return json.MarshalIndent(changes, "", "  ")
}
