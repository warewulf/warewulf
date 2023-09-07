package set

import (
	"log"

	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/kernel"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/spf13/cobra"
)

type variables struct {
	setNetDevDel string
	setDiskDel   string
	setPartDel   string
	setFsDel     string
	setNodeAll   bool
	setYes       bool
	setForce     bool
	netName      string
	partName     string
	diskName     string
	fsName       string
	profileConf  node.NodeConf
	converters   []func() error
}

func GetCommand() *cobra.Command {
	vars := variables{}
	vars.profileConf = node.NewConf()

	baseCmd := &cobra.Command{
		Use:   "set [OPTIONS] [PROFILE ...]",
		Short: "Configure node profile properties",
		Long: "This command sets configuration properties for the node PROFILE(s).\n\n" +
			"Note: use the string 'UNSET' to remove a configuration",
		Args: cobra.MinimumNArgs(0),
		RunE: CobraRunE(&vars),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			nodeDB, _ := node.New()
			profiles, _ := nodeDB.FindAllProfiles()
			var p_names []string
			for _, profile := range profiles {
				p_names = append(p_names, profile.Id.Get())
			}
			return p_names, cobra.ShellCompDirectiveNoFileComp
		},
	}
	vars.profileConf = node.NewConf()
	vars.converters = vars.profileConf.CreateFlags(baseCmd,
		[]string{"ipaddr", "ipaddr6", "ipmiaddr", "profile"})
	baseCmd.PersistentFlags().StringVar(&vars.netName, "netname", "default", "Set network name for network options")
	baseCmd.PersistentFlags().StringVarP(&vars.setNetDevDel, "netdel", "D", "", "Delete the node's network device")
	baseCmd.PersistentFlags().BoolVarP(&vars.setNodeAll, "all", "a", false, "Set all nodes")
	baseCmd.PersistentFlags().BoolVarP(&vars.setYes, "yes", "y", false, "Set 'yes' to all questions asked")
	baseCmd.PersistentFlags().BoolVarP(&vars.setForce, "force", "f", false, "Force configuration (even on error)")
	baseCmd.PersistentFlags().StringVar(&vars.fsName, "fsname", "", "set the file system name which must match a partition name")
	baseCmd.PersistentFlags().StringVar(&vars.partName, "partname", "", "set the partition name so it can be used by a file system")
	baseCmd.PersistentFlags().StringVar(&vars.diskName, "diskname", "", "set disk device name for the partition")
	baseCmd.PersistentFlags().StringVar(&vars.setDiskDel, "diskdel", "", "delete the disk from the configuration")
	baseCmd.PersistentFlags().StringVar(&vars.setPartDel, "partdel", "", "delete the partition from the configuration")
	baseCmd.PersistentFlags().StringVar(&vars.setFsDel, "fsdel", "", "delete the  from the configuration")
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
	return baseCmd
}
