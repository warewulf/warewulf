package list

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/api/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {

	nodeInfo, err := node.NodeList(args)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		return
	}

	if ShowAll {
		for i := 0; i < len(nodeInfo); i++ {
			ni := nodeInfo[i]
			nodeName := ni.Id.Value

			fmt.Printf("################################################################################\n")
			fmt.Printf("%-20s %-18s %-12s %s\n", "NODE", "FIELD", "PROFILE", "VALUE")

			fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, "Id", ni.Id.Source, ni.Id.Print)
			fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, "Comment", ni.Comment.Source, ni.Comment.Print)
			fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, "Cluster", ni.Cluster.Source, ni.Cluster.Print)
			fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, "Profiles", "--", strings.Join(ni.Profiles, ",'"))

			fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, "Discoverable", ni.Discoverable.Source, ni.Discoverable.Print)

			fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, "Container", ni.Container.Source, ni.Container.Print)
			fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, "KernelOverride", ni.KernelOverride.Source, ni.KernelOverride.Print)
			fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, "KernelArgs", ni.KernelArgs.Source, ni.KernelArgs.Print)
			fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, "SystemOverlay", ni.SystemOverlay.Source, ni.SystemOverlay.Print)
			fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, "RuntimeOverlay", ni.RuntimeOverlay.Source, ni.RuntimeOverlay.Print)
			fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, "Ipxe", ni.Ipxe.Source, ni.Ipxe.Print)
			fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, "Init", ni.Init.Source, ni.Init.Print)
			fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, "Root", ni.Root.Source, ni.Root.Print)
			fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, "AssetKey", ni.AssetKey.Source, ni.AssetKey.Print)

			fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, "IpmiIpaddr", ni.IpmiIpaddr.Source, ni.IpmiIpaddr.Print)
			fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, "IpmiNetmask", ni.IpmiNetmask.Source, ni.IpmiNetmask.Print)
			fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, "IpmiPort", ni.IpmiPort.Source, ni.IpmiPort.Print)
			fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, "IpmiGateway", ni.IpmiGateway.Source, ni.IpmiGateway.Print)
			fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, "IpmiUserName", ni.IpmiUserName.Source, ni.IpmiUserName.Print)
			fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, "IpmiInterface", ni.IpmiInterface.Source, ni.IpmiInterface.Print)

			for keyname, key := range ni.Tags {
				fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, "Tag["+keyname+"]", key.Source, key.Print)
			}

			for name, netdev := range ni.NetDevs {
				fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, name+":DEVICE", netdev.Device.Source, netdev.Device.Print)
				fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, name+":HWADDR", netdev.Hwaddr.Source, netdev.Hwaddr.Print)
				fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, name+":IPADDR", netdev.Ipaddr.Source, netdev.Ipaddr.Print)
				fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, name+":NETMASK", netdev.Netmask.Source, netdev.Netmask.Print)
				fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, name+":GATEWAY", netdev.Gateway.Source, netdev.Gateway.Print)
				fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, name+":TYPE", netdev.Type.Source, netdev.Type.Print)
				fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, name+":ONBOOT", netdev.Onboot.Source, netdev.Onboot.Print)
				fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, name+":DEFAULT", netdev.Primary.Source, netdev.Primary.Print)
			}
		}
	} else if ShowNet {
		fmt.Printf("%-22s %-6s %-18s %-15s %-15s\n", "NODE NAME", "DEVICE", "HWADDR", "IPADDR", "GATEWAY")
		fmt.Println(strings.Repeat("=", 80))

		for i := 0; i < len(nodeInfo); i++ {
			ni := nodeInfo[i]
			nodeName := ni.Id.Value

			if len(ni.NetDevs) > 0 {
				for name, dev := range ni.NetDevs {
					fmt.Printf("%-22s %-6s %-18s %-15s %-15s\n", nodeName, name, dev.Hwaddr.Print, dev.Ipaddr.Print, dev.Gateway.Print)
				}
			} else {
				fmt.Printf("%-22s %-6s %-18s %-15s %-15s\n", nodeName, "--", "--", "--", "--")
			}
		}
	} else if ShowIpmi {
		fmt.Printf("%-22s %-16s %-10s %-20s %-20s %-14s\n", "NODE NAME", "IPMI IPADDR", "IPMI PORT", "IPMI USERNAME", "IPMI PASSWORD", "IPMI INTERFACE")
		fmt.Println(strings.Repeat("=", 108))

		for i := 0; i < len(nodeInfo); i++ {
			ni := nodeInfo[i]
			nodeName := ni.Id.Value

			fmt.Printf("%-22s %-16s %-10s %-20s %-20s %-14s\n", nodeName, ni.IpmiIpaddr.Print, ni.IpmiPort.Print, ni.IpmiUserName.Print, ni.IpmiPassword.Print, ni.IpmiInterface.Print)
		}

	} else if ShowLong {
		fmt.Printf("%-22s %-26s %-35s %s\n", "NODE NAME", "KERNEL OVERRIDE", "CONTAINER", "OVERLAYS (S/R)")
		fmt.Println(strings.Repeat("=", 120))

		for i := 0; i < len(nodeInfo); i++ {
			ni := nodeInfo[i]
			nodeName := ni.Id.Value

			fmt.Printf("%-22s %-26s %-35s %s\n", nodeName, ni.KernelOverride.Print, ni.Container.Print, ni.SystemOverlay.Print+"/"+ni.RuntimeOverlay.Print)
		}

	} else {
		fmt.Printf("%-22s %-26s %s\n", "NODE NAME", "PROFILES", "NETWORK")
		fmt.Println(strings.Repeat("=", 80))

		for i := 0; i < len(nodeInfo); i++ {
			ni := nodeInfo[i]
			nodeName := ni.Id.Value

			var netdevs []string
			if len(ni.NetDevs) > 0 {
				for name, dev := range ni.NetDevs {
					netdevs = append(netdevs, fmt.Sprintf("%s:%s", name, dev.Ipaddr.Print))
				}
			}
			sort.Strings(netdevs)
			fmt.Printf("%-22s %-26s %s\n", nodeName, strings.Join(ni.Profiles, ","), strings.Join(netdevs, ", "))
		}
	}
	return
}
