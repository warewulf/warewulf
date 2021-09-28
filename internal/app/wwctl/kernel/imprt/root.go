package imprt

import (
	"log"

	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:   "import [flags] [kernel version]",
		Short: "Import Kernel version into Warewulf",
		Long: "This will import a Kernel version from the control node into Warewulf for nodes\n" +
			"to be configured to boot on.",
		RunE: CobraRunE,
		Args: cobra.MinimumNArgs(1),
	}
	BuildAll     bool
	ByNode       bool
	SetDefault   bool
	OptRoot      string
	OptContainer string
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
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
