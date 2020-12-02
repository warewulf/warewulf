package list

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
	"sort"
	"strings"
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

	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Gid.Get() < nodes[j].Gid.Get() {
			return true
		} else if nodes[i].Gid.Get() == nodes[j].Gid.Get() {
			if nodes[i].Id.Get() < nodes[j].Id.Get() {
				return true
			}
		}
		return false
	})

	if ShowAll == true {
		for _, node := range nodes {
			fmt.Printf("################################################################################\n")
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "Id", node.Id.Source(), node.Id.Print())
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "Comment", node.Comment.Source(), node.Comment.Print())
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "GroupName", node.Gid.Source(), node.Gid.Print())
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "DomainName", node.DomainName.Source(), node.DomainName.Print())
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "Profiles (Group)", "group", strings.Join(node.GroupProfiles, ","))
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "Profiles (Node)", "node", strings.Join(node.Profiles, ","))

			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "Vnfs", node.Vnfs.Source(), node.Vnfs.Print())
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "KernelVersion", node.KernelVersion.Source(), node.KernelVersion.Print())
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "KernelArgs", node.KernelArgs.Source(), node.KernelArgs.Print())
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "RuntimeOverlay", node.RuntimeOverlay.Source(), node.RuntimeOverlay.Print())
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "SystemOverlay", node.SystemOverlay.Source(), node.SystemOverlay.Print())
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "Ipxe", node.Ipxe.Source(), node.Ipxe.Print())
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "IpmiIpaddr", node.IpmiIpaddr.Source(), node.IpmiIpaddr.Print())
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "IpmiNetmask", node.IpmiNetmask.Source(), node.IpmiNetmask.Print())
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "IpmiUserName", node.IpmiUserName.Source(), node.IpmiUserName.Print())

			for name, netdev := range node.NetDevs {
				fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), name+":IPADDR", "node", netdev.Ipaddr)
				fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), name+":NETMASK", "node", netdev.Netmask)
				fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), name+":GATEWAY", "node", netdev.Gateway)
				fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), name+":HWADDR", "node", netdev.Hwaddr)
			}

			//			v := reflect.ValueOf(node)
			//			typeOfS := v.Type()
			//			for i := 0; i< v.NumField(); i++ {
			//				//TODO: Fix for NetDevs and Interface should print Fprint() method
			//				fmt.Printf("%-25s %s = %#v\n", node.Fqdn.Get(), typeOfS.Field(i).Name, v.Field(i).Interface())
			//			}
		}

	} else if ShowNet == true {
		fmt.Printf("%-22s %-6s %-18s %-15s %-15s\n", "NODE NAME", "DEVICE", "HWADDR", "IPADDR", "GATEWAY")
		fmt.Println(strings.Repeat("=", 80))

		for _, node := range nodes {
			if len(node.NetDevs) > 0 {
				for name, dev := range node.NetDevs {
					fmt.Printf("%-22s %-6s %-18s %-15s %-15s\n", node.Fqdn.Get(), name, dev.Hwaddr, dev.Ipaddr, dev.Gateway)
				}
			} else {
				fmt.Printf("%-22s %-6s %-18s %-15s %-15s\n", node.Fqdn.Get(), "--", "--", "--", "--")
			}
		}

	} else if ShowIpmi == true {
		fmt.Printf("%-22s %-16s %-20s %-20s\n", "NODE NAME", "IPMI IPADDR", "IPMI USERNAME", "IPMI PASSWORD")
		fmt.Println(strings.Repeat("=", 80))

		for _, node := range nodes {
			fmt.Printf("%-22s %-16s %-20s %-20s\n", node.Fqdn.Get(), node.IpmiIpaddr.Print(), node.IpmiUserName.Print(), node.IpmiPassword.Print())
		}

	} else if ShowLong == true {
		fmt.Printf("%-22s %-12s %-26s %-35s %s\n", "NODE NAME", "GROUP NAME", "KERNEL VERSION", "VNFS IMAGE", "OVERLAYS (S/R)")
		fmt.Println(strings.Repeat("=", 120))

		for _, node := range nodes {
			fmt.Printf("%-22s %-12s %-26s %-35s %s\n", node.Fqdn.Get(), node.Gid.Print(), node.KernelVersion.Print(), node.Vnfs.Print(), node.SystemOverlay.Print()+"/"+node.RuntimeOverlay.Print())
		}

	} else {
		fmt.Printf("%-22s: %-25s %s\n", "NODE NAME", "PROFILES", "NETDEVS")

		for _, node := range nodes {
			var netdevs []string
			if len(node.NetDevs) > 0 {
				for name, dev := range node.NetDevs {
					netdevs = append(netdevs, fmt.Sprintf("%s[%s/%s]", name, dev.Ipaddr, dev.Netmask))
				}
			}
			sort.Strings(netdevs)

			allProfiles := append(node.GroupProfiles, node.Profiles...)
			fmt.Printf("%-22s: %-25s %s\n", node.Fqdn.Get(), strings.Join(allProfiles, ","), strings.Join(netdevs, ", "))
		}

	}

	return nil
}
