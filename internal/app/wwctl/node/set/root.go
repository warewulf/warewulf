package set

import (
	"fmt"
	"log"

	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/kernel"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "set [OPTIONS] PATTERN [PATTERN ...]",
		Short:                 "Configure node properties",
		Long:                  "This command sets configuration properties for nodes matching PATTERN.\n\nNote: use the string 'UNSET' to remove a configuration",
		Args:                  cobra.MinimumNArgs(0),
		RunE:                  CobraRunE,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			nodeDB, _ := node.New()
			nodes, _ := nodeDB.FindAllNodes()
			var node_names []string
			for _, node := range nodes {
				node_names = append(node_names, node.Id.Get())
			}
			return node_names, cobra.ShellCompDirectiveNoFileComp
		},
	}
	SetKernelOverride string
	SetNetName        string
	SetNetDev         string
	SetIpaddr         string
	SetNetmask        string
	SetGateway        string
	SetHwaddr         string
	SetType           string
	SetNetOnBoot      string
	SetNetPrimary     string
	SetNetDevDel      bool
	SetClusterName    string
	SetIpxe           string
	SetInitOverlay    string
	SetRuntimeOverlay []string
	SetSystemOverlay  []string
	SetIpmiIpaddr     string
	SetIpmiNetmask    string
	SetIpmiPort       string
	SetIpmiGateway    string
	SetIpmiUsername   string
	SetIpmiPassword   string
	SetIpmiInterface  string
	SetIpmiWrite      string
	SetNodeAll        bool
	SetYes            bool
	SetProfile        string
	SetAddProfile     []string
	SetDelProfile     []string
	SetForce          bool
	SetDiscoverable   bool
	SetUndiscoverable bool
	SetTags           []string
	SetDelTags        []string
	SetNetTags        []string
	SetNetDelTags     []string
	OptionStrMap      map[string]*string
	KernelStrMapmap   map[string]*string
)

func init() {
	//var emptyNodeConf node.NodeConf
	//var emptyKernelE node.KernelEntry
	myBase := node.CobraCommand{baseCmd}
	var emptyNodeConf node.NodeConf
	emptyNodeConf.Kernel = new(node.KernelConf)
	emptyNodeConf.Ipmi = new(node.IpmiConf)
	//emptyNodeConf.NetDevs = make(map[string]*node.NetDevs)
	OptionStrMap = myBase.CreateFlags(emptyNodeConf)

	fmt.Printf("OptionStrMap: %v\n", OptionStrMap)
	if err := baseCmd.RegisterFlagCompletionFunc("container", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		list, _ := container.ListSources()
		return list, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		log.Println(err)
	}
	baseCmd.PersistentFlags().StringVarP(&SetKernelOverride, "kerneloverride", "K", "", "Set kernel override version for nodes")
	if err := baseCmd.RegisterFlagCompletionFunc("kerneloverride", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		list, _ := kernel.ListKernels()
		return list, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		log.Println(err)
	}
	baseCmd.PersistentFlags().StringVarP(&SetClusterName, "cluster", "c", "", "Set the node's cluster group")
	baseCmd.PersistentFlags().StringVar(&SetIpxe, "ipxe", "", "Set the node's iPXE template name")
	baseCmd.PersistentFlags().StringVarP(&SetInitOverlay, "wwinit", "O", "", "Set the node's initialization overlay")
	baseCmd.PersistentFlags().StringSliceVarP(&SetRuntimeOverlay, "runtime", "R", []string{}, "Set the node's runtime overlay")
	if err := baseCmd.RegisterFlagCompletionFunc("runtime", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		list, _ := overlay.FindOverlays()
		return list, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		log.Println(err)
	}
	baseCmd.PersistentFlags().StringSliceVarP(&SetSystemOverlay, "system", "S", []string{}, "Set the node's system overlay")
	if err := baseCmd.RegisterFlagCompletionFunc("system", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		list, _ := overlay.FindOverlays()
		return list, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		log.Println(err)
	}
	baseCmd.PersistentFlags().StringVar(&SetIpmiIpaddr, "ipmiaddr", "", "Set the node's IPMI IP address")
	baseCmd.PersistentFlags().StringVar(&SetIpmiNetmask, "ipminetmask", "", "Set the node's IPMI netmask")
	baseCmd.PersistentFlags().StringVar(&SetIpmiPort, "ipmiport", "", "Set the node's IPMI port")
	baseCmd.PersistentFlags().StringVar(&SetIpmiGateway, "ipmigateway", "", "Set the node's IPMI gateway")
	baseCmd.PersistentFlags().StringVar(&SetIpmiUsername, "ipmiuser", "", "Set the node's IPMI username")
	baseCmd.PersistentFlags().StringVar(&SetIpmiPassword, "ipmipass", "", "Set the node's IPMI password")
	baseCmd.PersistentFlags().StringVar(&SetIpmiInterface, "ipmiinterface", "", "Set the node's IPMI interface (defaults: 'lan')")
	baseCmd.PersistentFlags().StringVar(&SetIpmiWrite, "ipmiwrite", "", "Enable/disable the write of impi configuration (yes/no)")
	baseCmd.PersistentFlags().StringSliceVar(&SetAddProfile, "addprofile", []string{}, "Add Profile(s) to node")
	baseCmd.PersistentFlags().StringSliceVar(&SetDelProfile, "delprofile", []string{}, "Remove Profile(s) to node")
	baseCmd.PersistentFlags().StringVarP(&SetProfile, "profile", "P", "", "Set the node's profile members (comma separated)")
	if err := baseCmd.RegisterFlagCompletionFunc("profile", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var list []string
		nodeDB, _ := node.New()
		profiles, _ := nodeDB.FindAllProfiles()
		for _, profile := range profiles {
			list = append(list, profile.Id.Get())
		}
		return list, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		log.Println(err)
	}
	baseCmd.PersistentFlags().StringVarP(&SetNetName, "netname", "n", "default", "Define the network name to configure")
	baseCmd.PersistentFlags().StringVarP(&SetNetDev, "netdev", "N", "", "Set the node's network device")
	baseCmd.PersistentFlags().StringVarP(&SetIpaddr, "ipaddr", "I", "", "Set the node's network device IP address")
	baseCmd.PersistentFlags().StringVarP(&SetNetmask, "netmask", "M", "", "Set the node's network device netmask")
	baseCmd.PersistentFlags().StringVarP(&SetGateway, "gateway", "G", "", "Set the node's network device gateway")
	baseCmd.PersistentFlags().StringVarP(&SetHwaddr, "hwaddr", "H", "", "Set the node's network device HW address")
	baseCmd.PersistentFlags().StringVarP(&SetType, "type", "T", "", "Set the node's network device type")
	baseCmd.PersistentFlags().StringVar(&SetNetOnBoot, "onboot", "", "Enable/disable device (yes/no)")
	baseCmd.PersistentFlags().StringVar(&SetNetPrimary, "primary", "", "Enable/disable device as primary (yes/no)")

	baseCmd.PersistentFlags().BoolVar(&SetNetDevDel, "netdel", false, "Delete the node's network device")
	baseCmd.PersistentFlags().StringSliceVar(&SetNetTags, "nettag", []string{}, "Define custom tag to network device (key=value)")
	baseCmd.PersistentFlags().StringSliceVar(&SetNetDelTags, "netdeltag", []string{}, "Delete tag for network device")

	baseCmd.PersistentFlags().StringSliceVarP(&SetTags, "tag", "t", []string{}, "Define custom tag (key=value)")
	baseCmd.PersistentFlags().StringSliceVar(&SetDelTags, "tagdel", []string{}, "Delete tag")

	baseCmd.PersistentFlags().BoolVarP(&SetNodeAll, "all", "a", false, "Set all nodes")

	baseCmd.PersistentFlags().BoolVarP(&SetYes, "yes", "y", false, "Set 'yes' to all questions asked")
	baseCmd.PersistentFlags().BoolVarP(&SetForce, "force", "f", false, "Force configuration (even on error)")
	baseCmd.PersistentFlags().BoolVar(&SetDiscoverable, "discoverable", false, "Make this node discoverable")
	baseCmd.PersistentFlags().BoolVar(&SetUndiscoverable, "undiscoverable", false, "Remove the discoverable flag")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
