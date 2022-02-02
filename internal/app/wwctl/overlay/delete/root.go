package delete

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "delete [OPTIONS] OVERLAY_NAME [FILE [FILE ...]]",
		Short:                 "Delete Warewulf Overlay or files",
		Long:                  "This command will delete FILEs within OVERLAY_NAME or the entire OVERLAY_NAME if no\nfiles are listed. Use with caution!",
		RunE:                  CobraRunE,
		Args:                  cobra.RangeArgs(1, 2),
		Aliases:               []string{"rm", "del"},
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
