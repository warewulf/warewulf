package set

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:                "set",
		Short:              "Set group configurations",
		Long:               "Set group configurations ",
		RunE:				CobraRunE,
	}
	SetVnfs string
	SetKernel string
	SetDomainName string
	SetIpxe string
	SetRuntimeOverlay string
	SetSystemOverlay string
	SetHostname string
	SetClearNodes bool
	SetIpmiIpaddr string
	SetIpmiUsername string
	SetIpmiPassword string
	SetGroupAll bool
)

func init() {
	baseCmd.PersistentFlags().StringVarP(&SetVnfs, "vnfs", "V", "", "Set node Virtual Node File System (VNFS)")
	baseCmd.PersistentFlags().StringVarP(&SetKernel, "kernel", "K", "", "Set Kernel version for nodes")
	baseCmd.PersistentFlags().StringVarP(&SetDomainName, "domain", "D", "", "Set the node's domain name")
	baseCmd.PersistentFlags().StringVarP(&SetIpxe, "ipxe", "P", "", "Set the node's iPXE template name")
	baseCmd.PersistentFlags().StringVarP(&SetRuntimeOverlay, "runtime", "R", "", "Set the node's runtime overlay")
	baseCmd.PersistentFlags().StringVarP(&SetSystemOverlay, "system", "S", "", "Set the node's system overlay")
	baseCmd.PersistentFlags().StringVar(&SetIpmiIpaddr, "ipmi", "", "Set the node's IPMI address")
	baseCmd.PersistentFlags().StringVar(&SetIpmiUsername, "ipmiuser", "", "Set the node's IPMI username")
	baseCmd.PersistentFlags().StringVar(&SetIpmiPassword, "ipmipass", "", "Set the node's IPMI password")

	baseCmd.PersistentFlags().BoolVarP(&SetClearNodes, "clear", "c", false, "Clear node configurations when setting parent group")
	baseCmd.PersistentFlags().BoolVarP(&SetGroupAll, "all", "a", false, "Set all nodes")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
