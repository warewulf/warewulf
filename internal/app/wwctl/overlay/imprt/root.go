package imprt

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "import [OPTIONS] OVERLAY_NAME FILE [NEW_NAME]",
		Short:                 "Import a file into a Warewulf Overlay",
		Long:                  "This command imports the FILE into the Warewulf OVERLAY_NAME.\nOptionally, the file can be renamed to NEW_NAME",
		RunE:                  CobraRunE,
		Args:                  cobra.RangeArgs(2, 3),
		Aliases:               []string{"cp"},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completions.Overlays(cmd, args, toComplete)
			} else if len(args) == 1 {
				return completions.LocalFiles(cmd, args, toComplete)
			} else {
				return completions.None(cmd, args, toComplete)
			}
		},
	}
	OverwriteFile bool
	CreateDirs    bool
	Workers       int
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&OverwriteFile, "overwrite", "o", false, "Overwrite file if exists")
	baseCmd.PersistentFlags().BoolVarP(&CreateDirs, "parents", "p", false, "Create any necessary parent directories")
	baseCmd.PersistentFlags().IntVar(&Workers, "workers", 0, "The number of parallel workers building overlays (<=0 indicates 1 worker per CPU)")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
