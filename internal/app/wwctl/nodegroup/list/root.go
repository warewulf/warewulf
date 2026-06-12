package list

import (
	"github.com/spf13/cobra"

	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

// GetCommand returns the `wwctl nodegroup list` cobra command.
func GetCommand() *cobra.Command {
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "list [NODEGROUP ...]",
		Short:                 "List node groups and their members",
		Long: "Show nodegroups and their members.\n" +
			"Membership is the union of the top-level `nodegroups:` stanza, the\n" +
			"per-node `nodegroups:` field, and any `nodegroups:` declared on a\n" +
			"profile the node inherits. Without the NODEGROUP argument, all\n" +
			"groups are shown.",
		RunE:              CobraRunE(),
		Aliases:           []string{"ls"},
		Args:              cobra.ArbitraryArgs,
		ValidArgsFunction: completions.Nodegroups,
	}
	return baseCmd
}
