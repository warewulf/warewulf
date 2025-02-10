package delete

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "delete [OPTIONS] NODE [NODE ...]",
		Short:                 "Delete a node from Warewulf",
		Long:                  "This command will remove NODE(s) from the Warewulf node configuration.",
		Args:                  cobra.MinimumNArgs(1),
		RunE:                  CobraRunE,
		Aliases:               []string{"rm", "del", "remove"},
		ValidArgsFunction:     completions.Nodes,
	}
	SetYes   bool
	SetForce bool // no hash checking, so always using force
)

func init() {
	SetForce = true
	baseCmd.PersistentFlags().BoolVarP(&SetYes, "yes", "y", false, "Set 'yes' to all questions asked")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
