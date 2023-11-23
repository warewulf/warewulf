package add

import (
	"log"

	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/kernel"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/spf13/cobra"
)

type variables struct {
	netName      string
	profileConf  node.NodeConf
	SetNetDevDel string
	SetNodeAll   bool
	SetYes       bool
	SetForce     bool
	fsName       string
	partName     string
	diskName     string
	ProfileConf  node.NodeConf
	Converters   []func() error
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	vars := variables{}
	vars.profileConf = node.NewConf()
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "add PROFILE",
		Short:                 "Add a new node profile",
		Long:                  "This command adds a new named PROFILE.",
		RunE:                  CobraRunE(&vars),
		Args:                  cobra.ExactArgs(1),
	}
	vars.Converters = vars.profileConf.CreateFlags(baseCmd,
		[]string{"ipaddr", "ipaddr6", "ipmiaddr", "profile"})
	baseCmd.PersistentFlags().StringVar(&vars.netName, "netname", "", "Set network name for network options")
	baseCmd.PersistentFlags().StringVar(&vars.fsName, "fsname", "", "set the file system name which must match a partition name")
	baseCmd.PersistentFlags().StringVar(&vars.partName, "partname", "", "set the partition name so it can be used by a file system")
	baseCmd.PersistentFlags().StringVar(&vars.diskName, "diskname", "", "set disk device name for the partition")
	// register the command line completions
	if err := baseCmd.RegisterFlagCompletionFunc("container", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		list, _ := container.ListSources()
		return list, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		log.Println(err)
	}
	if err := baseCmd.RegisterFlagCompletionFunc("kerneloverride", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		list, _ := kernel.ListKernels()
		return list, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		log.Println(err)
	}
	if err := baseCmd.RegisterFlagCompletionFunc("runtime", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		list, _ := overlay.FindOverlays()
		return list, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		log.Println(err)
	}
	if err := baseCmd.RegisterFlagCompletionFunc("wwinit", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		list, _ := overlay.FindOverlays()
		return list, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		log.Println(err)
	}
	baseCmd.PersistentFlags().BoolVarP(&vars.SetYes, "yes", "y", false, "Set 'yes' to all questions asked")
	return baseCmd
}
