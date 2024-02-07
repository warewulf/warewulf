package exec

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/container/exec/child"
	"github.com/warewulf/warewulf/internal/pkg/container"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "exec [OPTIONS] CONTAINER COMMAND",
		Short:                 "Run a command inside of a Warewulf container",
		Long: "Run a COMMAND inside of a warewulf CONTAINER.\n" +
			"This is commonly used with an interactive shell such as /bin/bash\n" +
			"to run a virtual environment within the container.",
		RunE: CobraRunE,
		Args: cobra.MinimumNArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			list, _ := container.ListSources()
			return list, cobra.ShellCompDirectiveNoFileComp
		},
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	}
	SyncUser   bool
	binds      []string
	overlayDir string
	nodeName   string
)

func init() {
	baseCmd.AddCommand(child.GetCommand())
	baseCmd.PersistentFlags().StringArrayVarP(&binds, "bind", "b", []string{}, "Bind a local path into the container (must exist)")
	baseCmd.PersistentFlags().BoolVar(&SyncUser, "syncuser", false, "Synchronize UIDs/GIDs from host to container")
	baseCmd.PersistentFlags().StringVar(&overlayDir, "overlaydir", "", "Use tempdir for constructing the overlay fs")
	baseCmd.PersistentFlags().StringVarP(&nodeName, "node", "n", "", "Create a read only view of the container for the given node")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
