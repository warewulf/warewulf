package list

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
	"sort"
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

	sort.Slice(profiles, func(i, j int) bool {
		if profiles[i].Id.Get() < profiles[j].Id.Get() {
			return true
		}
		return false
	})

	for _, profile := range profiles {
		fmt.Printf("################################################################################\n")
		fmt.Printf("%-20s %-18s %s\n", "PROFILE NAME", "FIELD", "VALUE")
		fmt.Printf("%-20s %-18s %s\n", profile.Id.Get(), "Id", profile.Id.Print())
		fmt.Printf("%-20s %-18s %s\n", profile.Id.Get(), "Comment", profile.Comment.Print())
		fmt.Printf("%-20s %-18s %s\n", profile.Id.Get(), "ClusterName", profile.ClusterName.Print())

		fmt.Printf("%-20s %-18s %s\n", profile.Id.Get(), "Vnfs", profile.Vnfs.Print())
		fmt.Printf("%-20s %-18s %s\n", profile.Id.Get(), "KernelVersion", profile.KernelVersion.Print())
		fmt.Printf("%-20s %-18s %s\n", profile.Id.Get(), "KernelArgs", profile.KernelArgs.Print())
		fmt.Printf("%-20s %-18s %s\n", profile.Id.Get(), "RuntimeOverlay", profile.RuntimeOverlay.Print())
		fmt.Printf("%-20s %-18s %s\n", profile.Id.Get(), "SystemOverlay", profile.SystemOverlay.Print())
		fmt.Printf("%-20s %-18s %s\n", profile.Id.Get(), "Ipxe", profile.Ipxe.Print())
		fmt.Printf("%-20s %-18s %s\n", profile.Id.Get(), "IpmiIpaddr", profile.IpmiIpaddr.Print())
		fmt.Printf("%-20s %-18s %s\n", profile.Id.Get(), "IpmiNetmask", profile.IpmiNetmask.Print())
		fmt.Printf("%-20s %-18s %s\n", profile.Id.Get(), "IpmiUserName", profile.IpmiUserName.Print())

		for name, netdev := range profile.NetDevs {
			fmt.Printf("%-20s %-18s %s\n", profile.Id.Get(), name+":IPADDR", netdev.Ipaddr.Print())
			fmt.Printf("%-20s %-18s %s\n", profile.Id.Get(), name+":NETMASK", netdev.Netmask.Print())
			fmt.Printf("%-20s %-18s %s\n", profile.Id.Get(), name+":GATEWAY", netdev.Gateway.Print())
			fmt.Printf("%-20s %-18s %s\n", profile.Id.Get(), name+":HWADDR", netdev.Hwaddr.Print())
			fmt.Printf("%-20s %-18s %s\n", profile.Id.Get(), name+":TYPE", netdev.Hwaddr.Print())

		}
	}

	return nil
}
