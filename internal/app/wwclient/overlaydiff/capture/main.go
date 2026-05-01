package capture

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/overlaydiff"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	format := strings.ToLower(OutputFormat)
	if format != "table" && format != "json" {
		return fmt.Errorf("invalid format %q: expected table or json", OutputFormat)
	}

	changes, err := overlaydiff.Diff(SourcePath, BaselinePath)
	if err != nil {
		return err
	}

	switch format {
	case "json":
		data, err := overlaydiff.FormatJSON(changes)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", data)
	default:
		_, _ = fmt.Fprint(cmd.OutOrStdout(), overlaydiff.FormatTable(changes))
	}

	return nil
}
