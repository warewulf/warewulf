package exec

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/image/exec/child"
	"github.com/warewulf/warewulf/internal/pkg/image"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "exec [OPTIONS] IMAGE COMMAND",
		Short:                 "Run a command inside of a Warewulf image",
		Long: "Run a COMMAND inside of a warewulf IMAGE.\n" +
			"This is commonly used with an interactive shell such as /bin/bash\n" +
			"to run a virtual environment within the image.",
		RunE: CobraRunE,
		Args: cobra.MinimumNArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			list, _ := image.ListSources()
			return list, cobra.ShellCompDirectiveNoFileComp
		},
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	}
	SyncUser bool
	Build    bool
	binds    []string
	nodeName string
)

func init() {
	baseCmd.AddCommand(child.GetCommand())
	baseCmd.PersistentFlags().StringArrayVarP(&binds, "bind", "b", []string{}, `source[:destination[:{ro|copy}]]
Bind a local path which must exist into the image. If destination is not
set, uses the same path as source. "ro" binds read-only. "copy" temporarily
copies the file into the image.`)
	baseCmd.PersistentFlags().BoolVar(&SyncUser, "syncuser", false, "Synchronize UIDs/GIDs from host to image")
	baseCmd.PersistentFlags().BoolVar(&Build, "build", true, "(Re)build the image automatically")
	baseCmd.PersistentFlags().StringVarP(&nodeName, "node", "n", "", "Create a read only view of the image for the given node")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
