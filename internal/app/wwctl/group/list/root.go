package list

import (
	"github.com/spf13/cobra"

	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

// GetCommand returns the `wwctl group list` cobra command.
func GetCommand() *cobra.Command {
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "list [GROUP ...]",
		Short:                 "List node groups and their members",
		Long: "Show groups and their members.\n" +
			"Membership is the union of the per-node `groups:` field and any\n" +
			"`groups:` declared on a profile the node inherits. Without the\n" +
			"GROUP argument, all groups are shown.",
		RunE:              CobraRunE(),
		Aliases:           []string{"ls"},
		Args:              cobra.ArbitraryArgs,
		ValidArgsFunction: completions.Groups,
	}
	return baseCmd
}
