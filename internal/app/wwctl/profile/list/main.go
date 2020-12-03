package list

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	profiles, err := nodeDB.FindAllProfiles()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not find all nodes: %s\n", err)
		os.Exit(1)
	}

	for _, node := range profiles {
		fmt.Printf("################################################################################\n")
		fmt.Printf("%-20s %-18s: %s\n", node.Id.Get(), "Id", node.Id.Print())
		fmt.Printf("%-20s %-18s: %s\n", node.Id.Get(), "Comment", node.Comment.Print())

		fmt.Printf("%-20s %-18s: %s\n", node.Id.Get(), "Vnfs", node.Vnfs.Print())
		fmt.Printf("%-20s %-18s: %s\n", node.Id.Get(), "KernelVersion", node.KernelVersion.Print())
		fmt.Printf("%-20s %-18s: %s\n", node.Id.Get(), "KernelArgs", node.KernelArgs.Print())
		fmt.Printf("%-20s %-18s: %s\n", node.Id.Get(), "RuntimeOverlay", node.RuntimeOverlay.Print())
		fmt.Printf("%-20s %-18s: %s\n", node.Id.Get(), "SystemOverlay", node.SystemOverlay.Print())
		fmt.Printf("%-20s %-18s: %s\n", node.Id.Get(), "Ipxe", node.Ipxe.Print())
		fmt.Printf("%-20s %-18s: %s\n", node.Id.Get(), "IpmiIpaddr", node.IpmiIpaddr.Print())
		fmt.Printf("%-20s %-18s: %s\n", node.Id.Get(), "IpmiNetmask", node.IpmiNetmask.Print())
		fmt.Printf("%-20s %-18s: %s\n", node.Id.Get(), "IpmiUserName", node.IpmiUserName.Print())

		for name, netdev := range node.NetDevs {
			fmt.Printf("%-20s %-18s: %s\n", node.Id.Get(), name+":IPADDR", netdev.Ipaddr.Get())
			fmt.Printf("%-20s %-18s: %s\n", node.Id.Get(), name+":NETMASK", netdev.Netmask.Get())
			fmt.Printf("%-20s %-18s: %s\n", node.Id.Get(), name+":GATEWAY", netdev.Gateway.Get())
			fmt.Printf("%-20s %-18s: %s\n", node.Id.Get(), name+":HWADDR", netdev.Hwaddr.Get())
			fmt.Printf("%-20s %-18s: %s\n", node.Id.Get(), name+":TYPE", netdev.Hwaddr.Get())

		}
	}

	return nil
}
