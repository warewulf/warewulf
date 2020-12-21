package set

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:   "set",
		Short: "Set profile configurations",
		Long:  "Profile configurations ",
		RunE:  CobraRunE,
	}
	SetAll            bool
	SetForce          bool
	SetComment        string
	SetContainer      string
	SetKernel         string
	SetClusterName    string
	SetIpxe           string
	SetRuntimeOverlay string
	SetSystemOverlay  string
	SetIpmiNetmask    string
	SetIpmiGateway    string
	SetIpmiUsername   string
	SetIpmiPassword   string
	SetNetDev         string
	SetIpaddr         string
	SetNetmask        string
	SetGateway        string
	SetHwaddr         string
	SetNetDevDel      bool
	SetInit           string
	SetRoot           string
)

func init() {
	baseCmd.PersistentFlags().StringVar(&SetComment, "comment", "", "Set a comment for this node")
	baseCmd.PersistentFlags().StringVarP(&SetContainer, "container", "C", "", "Set the container (VNFS) for this node")
	baseCmd.PersistentFlags().StringVarP(&SetKernel, "kernel", "K", "", "Set Kernel version for nodes")
	baseCmd.PersistentFlags().StringVarP(&SetClusterName, "cluster", "c", "", "Set the node's cluster group")
	baseCmd.PersistentFlags().StringVarP(&SetIpxe, "ipxe", "P", "", "Set the node's iPXE template name")
	baseCmd.PersistentFlags().StringVarP(&SetInit, "init", "i", "", "Define the init process to boot the container")
	baseCmd.PersistentFlags().StringVar(&SetRoot, "root", "", "Define the rootfs")

	baseCmd.PersistentFlags().StringVarP(&SetRuntimeOverlay, "runtime", "R", "", "Set the node's runtime overlay")
	baseCmd.PersistentFlags().StringVarP(&SetSystemOverlay, "system", "S", "", "Set the node's system overlay")
	baseCmd.PersistentFlags().StringVar(&SetIpmiNetmask, "ipminetmask", "", "Set the node's IPMI netmask")
	baseCmd.PersistentFlags().StringVar(&SetIpmiGateway, "ipmigateway", "", "Set the node's IPMI gateway")
	baseCmd.PersistentFlags().StringVar(&SetIpmiUsername, "ipmiuser", "", "Set the node's IPMI username")
	baseCmd.PersistentFlags().StringVar(&SetIpmiPassword, "ipmipass", "", "Set the node's IPMI password")

	baseCmd.PersistentFlags().StringVarP(&SetNetDev, "netdev", "N", "", "Define the network device to configure")
	baseCmd.PersistentFlags().StringVarP(&SetIpaddr, "ipaddr", "I", "", "Set the node's network device IP address")
	baseCmd.PersistentFlags().StringVarP(&SetNetmask, "netmask", "M", "", "Set the node's network device netmask")
	baseCmd.PersistentFlags().StringVarP(&SetGateway, "gateway", "G", "", "Set the node's network device gateway")
	baseCmd.PersistentFlags().StringVarP(&SetHwaddr, "hwaddr", "H", "", "Set the node's network device HW address")
	baseCmd.PersistentFlags().BoolVar(&SetNetDevDel, "netdel", false, "Delete the node's network device")

	baseCmd.PersistentFlags().BoolVarP(&SetAll, "all", "a", false, "Set all profiles")
	baseCmd.PersistentFlags().BoolVarP(&SetForce, "force", "f", false, "Force configuration (even on error)")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
