package list

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var err error
	var nodes []node.NodeInfo

	n, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	if len(args) > 0 {
		nodes, err = n.SearchByNameList(args)
	} else {
		nodes, err = n.FindAllNodes()
	}
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Cloud not get nodeList: %s\n", err)
		os.Exit(1)
	}

	if ShowAll {
		for _, node := range nodes {
			fmt.Printf("################################################################################\n")
			fmt.Printf("%-20s %-18s %-12s %s\n", "NODE", "FIELD", "PROFILE", "VALUE")
			fmt.Printf("%-20s %-18s %-12s %s\n", node.Id.Get(), "Id", node.Id.Source(), node.Id.Print())
			fmt.Printf("%-20s %-18s %-12s %s\n", node.Id.Get(), "Comment", node.Comment.Source(), node.Comment.Print())
			fmt.Printf("%-20s %-18s %-12s %s\n", node.Id.Get(), "ClusterName", node.ClusterName.Source(), node.ClusterName.Print())
			fmt.Printf("%-20s %-18s %-12s %s\n", node.Id.Get(), "Profiles", "--", strings.Join(node.Profiles, ","))

			fmt.Printf("%-20s %-18s %-12s %t\n", node.Id.Get(), "Discoverable", node.Discoverable.Source(), node.Discoverable.PrintB())

			fmt.Printf("%-20s %-18s %-12s %s\n", node.Id.Get(), "ContainerName", node.ContainerName.Source(), node.ContainerName.Print())
			fmt.Printf("%-20s %-18s %-12s %s\n", node.Id.Get(), "KernelVersion", node.KernelVersion.Source(), node.KernelVersion.Print())
			fmt.Printf("%-20s %-18s %-12s %s\n", node.Id.Get(), "KernelArgs", node.KernelArgs.Source(), node.KernelArgs.Print())
			fmt.Printf("%-20s %-18s %-12s %s\n", node.Id.Get(), "RuntimeOverlay", node.RuntimeOverlay.Source(), node.RuntimeOverlay.Print())
			fmt.Printf("%-20s %-18s %-12s %s\n", node.Id.Get(), "SystemOverlay", node.SystemOverlay.Source(), node.SystemOverlay.Print())
			fmt.Printf("%-20s %-18s %-12s %s\n", node.Id.Get(), "Ipxe", node.Ipxe.Source(), node.Ipxe.Print())
			fmt.Printf("%-20s %-18s %-12s %s\n", node.Id.Get(), "Init", node.Init.Source(), node.Init.Print())
			fmt.Printf("%-20s %-18s %-12s %s\n", node.Id.Get(), "Root", node.Root.Source(), node.Root.Print())
			fmt.Printf("%-20s %-18s %-12s %s\n", node.Id.Get(), "IpmiIpaddr", node.IpmiIpaddr.Source(), node.IpmiIpaddr.Print())
			fmt.Printf("%-20s %-18s %-12s %s\n", node.Id.Get(), "IpmiNetmask", node.IpmiNetmask.Source(), node.IpmiNetmask.Print())
			fmt.Printf("%-20s %-18s %-12s %s\n", node.Id.Get(), "IpmiGateway", node.IpmiGateway.Source(), node.IpmiGateway.Print())
			fmt.Printf("%-20s %-18s %-12s %s\n", node.Id.Get(), "IpmiUserName", node.IpmiUserName.Source(), node.IpmiUserName.Print())

			for name, netdev := range node.NetDevs {
				fmt.Printf("%-20s %-18s %-12s %s\n", node.Id.Get(), name+":HWADDR", netdev.Hwaddr.Source(), netdev.Hwaddr.Print())
				fmt.Printf("%-20s %-18s %-12s %s\n", node.Id.Get(), name+":IPADDR", netdev.Ipaddr.Source(), netdev.Ipaddr.Print())
				fmt.Printf("%-20s %-18s %-12s %s\n", node.Id.Get(), name+":NETMASK", netdev.Netmask.Source(), netdev.Netmask.Print())
				fmt.Printf("%-20s %-18s %-12s %s\n", node.Id.Get(), name+":GATEWAY", netdev.Gateway.Source(), netdev.Gateway.Print())
				fmt.Printf("%-20s %-18s %-12s %s\n", node.Id.Get(), name+":TYPE", netdev.Type.Source(), netdev.Type.Print())
				fmt.Printf("%-20s %-18s %-12s %t\n", node.Id.Get(), name+":DEFAULT", netdev.Default.Source(), netdev.Default.PrintB())
			}

			//			v := reflect.ValueOf(node)
			//			typeOfS := v.Type()
			//			for i := 0; i< v.NumField(); i++ {
			//				//TODO: Fix for NetDevs and Interface should print Fprint() method
			//				fmt.Printf("%-25s %s = %#v\n", node.Fqdn.Get(), typeOfS.Field(i).Name, v.Field(i).Interface())
			//			}
		}

	} else if ShowNet {
		fmt.Printf("%-22s %-6s %-18s %-15s %-15s\n", "NODE NAME", "DEVICE", "HWADDR", "IPADDR", "GATEWAY")
		fmt.Println(strings.Repeat("=", 80))

		for _, node := range nodes {
			if len(node.NetDevs) > 0 {
				for name, dev := range node.NetDevs {
					fmt.Printf("%-22s %-6s %-18s %-15s %-15s\n", node.Id.Get(), name, dev.Hwaddr.Print(), dev.Ipaddr.Print(), dev.Gateway.Print())
				}
			} else {
				fmt.Printf("%-22s %-6s %-18s %-15s %-15s\n", node.Id.Get(), "--", "--", "--", "--")
			}
		}

	} else if ShowIpmi {
		fmt.Printf("%-22s %-16s %-20s %-20s\n", "NODE NAME", "IPMI IPADDR", "IPMI USERNAME", "IPMI PASSWORD")
		fmt.Println(strings.Repeat("=", 80))

		for _, node := range nodes {
			fmt.Printf("%-22s %-16s %-20s %-20s\n", node.Id.Get(), node.IpmiIpaddr.Print(), node.IpmiUserName.Print(), node.IpmiPassword.Print())
		}

	} else if ShowLong {
		fmt.Printf("%-22s %-26s %-35s %s\n", "NODE NAME", "KERNEL VERSION", "CONTAINER", "OVERLAYS (S/R)")
		fmt.Println(strings.Repeat("=", 120))

		for _, node := range nodes {
			fmt.Printf("%-22s %-26s %-35s %s\n", node.Id.Get(), node.KernelVersion.Print(), node.ContainerName.Print(), node.SystemOverlay.Print()+"/"+node.RuntimeOverlay.Print())
		}

	} else {
		fmt.Printf("%-22s %-26s %s\n", "NODE NAME", "PROFILES", "NETWORK")

		for _, node := range nodes {
			var netdevs []string
			if len(node.NetDevs) > 0 {
				for name, dev := range node.NetDevs {
					netdevs = append(netdevs, fmt.Sprintf("%s:%s", name, dev.Ipaddr.Print()))
				}
			}
			sort.Strings(netdevs)

			fmt.Printf("%-22s %-26s %s\n", node.Id.Get(), strings.Join(node.Profiles, ","), strings.Join(netdevs, ", "))
		}

	}

	return nil
}
