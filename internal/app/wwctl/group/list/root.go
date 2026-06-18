package list

import (
	"github.com/spf13/cobra"

	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

// GetCommand returns the `wwctl group list` cobra command.
func GetCommand() *cobra.Command {
	var noHeader bool
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "list [GROUP ...]",
		Short:                 "List node groups and their members",
		Long: "Show groups and their members.\n" +
			"Membership is the union of the per-node `groups:` field and any\n" +
			"`groups:` declared on a profile the node inherits. Without the\n" +
			"GROUP argument, all groups are shown (including the built-in\n" +
			"`all` group).",
		RunE:              cobraRunE(&noHeader),
		Aliases:           []string{"ls"},
		Args:              cobra.ArbitraryArgs,
		ValidArgsFunction: completions.Groups,
	}
	baseCmd.Flags().BoolVarP(&noHeader, "noheader", "n", false,
		"Print a comma-separated list of member nodes with no header or table formatting (deduped across all requested groups). Requires at least one GROUP.")
	return baseCmd
}
