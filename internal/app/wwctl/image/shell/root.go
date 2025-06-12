package shell

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/image"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "shell [OPTIONS] IMAGE",
		Short:                 "Run a shell inside of a Warewulf image",
		Long:                  "Run a interactive shell inside of a warewulf IMAGE.\n",
		Aliases:               []string{"chroot"},
		RunE:                  CobraRunE,
		Args:                  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			list, _ := image.ListSources()
			return list, cobra.ShellCompDirectiveNoFileComp
		},
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	}
	binds    []string
	nodeName string
	syncUser bool
	build    bool
)

func init() {
	baseCmd.PersistentFlags().StringArrayVarP(&binds, "bind", "b", []string{}, `source[:destination[:{ro|copy}]]
Bind a local path which must exist into the image.
If destination is not set, uses the same path as
source. "ro" binds read-only. "copy" temporarily
copies the file into the image.`)
	baseCmd.PersistentFlags().StringVarP(&nodeName, "node", "n", "", `Create a read only view of the image for the given
node`)
	baseCmd.PersistentFlags().BoolVar(&syncUser, "syncuser", false, "Synchronize UIDs/GIDs from host to image")
	baseCmd.PersistentFlags().BoolVar(&build, "build", true, "(Re)build the image automatically")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
