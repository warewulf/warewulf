package list

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

// Holds the variables which are needed in CobraRunE
type variables struct {
	ListContents bool
	ListLong     bool
	ShowPath     bool
}

// GetCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	vars := variables{}
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "list [OPTIONS] OVERLAY_NAME",
		Short:                 "List Warewulf Overlays and files",
		Long:                  "This command displays information about all Warewulf overlays or the specified\nOVERLAY_NAME. It also supports listing overlay content information.",
		RunE:                  CobraRunE(&vars),
		Aliases:               []string{"ls"},
		ValidArgsFunction:     completions.Overlays,
		Args:                  cobra.ArbitraryArgs,
	}
	baseCmd.PersistentFlags().BoolVarP(&vars.ListContents, "all", "a", false, "List the contents of overlays")
	baseCmd.PersistentFlags().BoolVarP(&vars.ListLong, "long", "l", false, "List 'long' of all overlay contents")
	baseCmd.PersistentFlags().BoolVarP(&vars.ShowPath, "path", "p", false, "Show the absolute path to the overlay")

	return baseCmd
}
