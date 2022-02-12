package imprt

import (
	"log"

	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "import [OPTIONS] KERNEL",
		Short:                 "Import Kernel version into Warewulf",
		Long:                  "This will import a boot KERNEL version from the control node into Warewulf",
		RunE:                  CobraRunE,
		Args:                  cobra.MinimumNArgs(0),
	}
	BuildAll     bool
	ByNode       bool
	SetDefault   bool
	OptRoot      string
	OptContainer string
	OptDetect    bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&BuildAll, "all", "a", false, "Build all overlays (runtime and system)")
	baseCmd.PersistentFlags().BoolVarP(&ByNode, "node", "n", false, "Build overlay for a particular node(s)")
	baseCmd.PersistentFlags().BoolVar(&SetDefault, "setdefault", false, "Set this kernel for the default profile")
	baseCmd.PersistentFlags().StringVarP(&OptRoot, "root", "r", "/", "Import kernel from root (chroot) directory")
	baseCmd.PersistentFlags().StringVarP(&OptContainer, "container", "C", "", "Import kernel from container")
	err := baseCmd.RegisterFlagCompletionFunc("container", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		list, _ := container.ListSources()
		return list, cobra.ShellCompDirectiveNoFileComp
	})
	if err != nil {
		log.Println(err)
	}
	baseCmd.PersistentFlags().BoolVarP(&OptDetect, "detect", "D", false, "Try to detect the kernel version in an automated way, needs the -C or -r option")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
