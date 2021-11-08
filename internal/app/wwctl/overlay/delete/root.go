package delete

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:   "delete [flags] (overlay kind) (overlay name) [overlay file]",
		Short: "Delete Warewulf Overlay or files",
		Long: "This command will delete files within an overlay or an entire overlay if no\n" +
			"files are given to remove (use with caution).",
		RunE:    CobraRunE,
		Args:    cobra.RangeArgs(2, 3),
		Aliases: []string{"rm", "del"},
	}
	Force   bool
	Parents bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&Force, "force", "f", false, "Force deletion of a non-empty overlay")
	baseCmd.PersistentFlags().BoolVarP(&Parents, "parents", "p", false, "Remove empty parent directories")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
