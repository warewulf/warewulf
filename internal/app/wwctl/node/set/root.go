package set

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:                "set",
		Short:              "Set node configurations",
		Long:               "Set node configurations ",
		RunE:				CobraRunE,
	}
	SetVnfs string
	SetKernel string
	SetNetDev string
	SetIpaddr string
	SetNetmask string
	SetGateway string
	SetHwaddr string
	SetNetDevDel bool
	SetDomainName string
	SetIpxe string
	SetRuntimeOverlay string
	SetSystemOverlay string
	SetHostname string
	SetIpmiIpaddr string
	SetIpmiUsername string
	SetIpmiPassword string

)

func init() {
	baseCmd.PersistentFlags().StringVarP(&SetVnfs, "vnfs", "V", "", "Set node Virtual Node File System (VNFS)")
	baseCmd.PersistentFlags().StringVarP(&SetKernel, "kernel", "K", "", "Set Kernel version for nodes")
	baseCmd.PersistentFlags().StringVarP(&SetDomainName, "domain", "D", "", "Set the node's domain name")
	baseCmd.PersistentFlags().StringVarP(&SetIpxe, "ipxe", "P", "", "Set the node's iPXE template name")
	baseCmd.PersistentFlags().StringVarP(&SetRuntimeOverlay, "runtime", "R", "", "Set the node's runtime overlay")
	baseCmd.PersistentFlags().StringVarP(&SetSystemOverlay, "system", "S", "", "Set the node's system overlay")
	baseCmd.PersistentFlags().StringVarP(&SetHostname, "hostname", "H", "", "Set the node's hostname")
	baseCmd.PersistentFlags().StringVar(&SetIpmiIpaddr, "ipmi", "", "Set the node's IPMI address")
	baseCmd.PersistentFlags().StringVar(&SetIpmiUsername, "ipmiuser", "", "Set the node's IPMI username")
	baseCmd.PersistentFlags().StringVar(&SetIpmiPassword, "ipmipass", "", "Set the node's IPMI password")

	baseCmd.PersistentFlags().StringVarP(&SetNetDev, "netdev", "n", "", "Define the network device to configure")
	baseCmd.PersistentFlags().StringVarP(&SetIpaddr, "ipaddr", "I", "", "Set the node's network device IP address")
	baseCmd.PersistentFlags().StringVarP(&SetNetmask, "netmask", "M", "", "Set the node's network device netmask")
	baseCmd.PersistentFlags().StringVarP(&SetGateway, "gateway", "G", "", "Set the node's network device gateway")
	baseCmd.PersistentFlags().StringVarP(&SetHwaddr, "hwaddr", "N", "", "Set the node's network device HW address")
	baseCmd.PersistentFlags().BoolVar(&SetNetDevDel, "delete", false, "Delete the node's network device")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
