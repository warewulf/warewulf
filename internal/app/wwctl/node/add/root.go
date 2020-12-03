package add

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:   "add",
		Short: "Add new node",
		Long:  "Add new node ",
		RunE:  CobraRunE,
		Args:  cobra.MinimumNArgs(1),
	}
	SetGroup      string
	SetController string
	SetNetDev     string
	SetIpaddr     string
	SetNetmask    string
	SetGateway    string
	SetHwaddr     string
)

func init() {
	baseCmd.PersistentFlags().StringVarP(&SetGroup, "group", "g", "default", "Group to add nodes to")
	baseCmd.PersistentFlags().StringVarP(&SetController, "controller", "c", "localhost", "Controller to add nodes to")
	baseCmd.PersistentFlags().StringVarP(&SetNetDev, "netdev", "n", "eth0", "Define the network device to configure")
	baseCmd.PersistentFlags().StringVarP(&SetIpaddr, "ipaddr", "I", "", "Set the node's network device IP address")
	baseCmd.PersistentFlags().StringVarP(&SetNetmask, "netmask", "M", "", "Set the node's network device netmask")
	baseCmd.PersistentFlags().StringVarP(&SetGateway, "gateway", "G", "", "Set the node's network device gateway")
	baseCmd.PersistentFlags().StringVarP(&SetHwaddr, "hwaddr", "H", "", "Set the node's network device HW address")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
