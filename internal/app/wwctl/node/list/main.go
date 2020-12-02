package list

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
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

	if ShowAll == true {
		for _, node := range nodes {
			fmt.Printf("################################################################################\n")
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "Id", node.Id.Source(), node.Id.String())
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "Comment", node.Comment.Source(), node.Comment.String())
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "GroupName", node.Gid.Source(), node.Gid.String())
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "DomainName", node.DomainName.Source(), node.DomainName.String())
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "Profiles (Group)", "group", strings.Join(node.GroupProfiles, ","))
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "Profiles (Node)", "node", strings.Join(node.Profiles, ","))

			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "Vnfs", node.Vnfs.Source(), node.Vnfs.String())
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "KernelVersion", node.KernelVersion.Source(), node.KernelVersion.String())
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "KernelArgs", node.KernelArgs.Source(), node.KernelArgs.String())
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "RuntimeOverlay", node.RuntimeOverlay.Source(), node.RuntimeOverlay.String())
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "SystemOverlay", node.SystemOverlay.Source(), node.SystemOverlay.String())
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "Ipxe", node.Ipxe.Source(), node.Ipxe.String())
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "IpmiIpaddr", node.IpmiIpaddr.Source(), node.IpmiIpaddr.String())
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "IpmiNetmask", node.IpmiNetmask.Source(), node.IpmiNetmask.String())
			fmt.Printf("%-20s %-18s %8s: %s\n", node.Fqdn.Get(), "IpmiUserName", node.IpmiUserName.Source(), node.IpmiUserName.String())

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
			fmt.Printf("%-22s %-16s %-20s %-20s\n", node.Fqdn.Get(), node.IpmiIpaddr.String(), node.IpmiUserName.String(), node.IpmiPassword.String())
		}

	} else if ShowLong == true {
		fmt.Printf("%-22s %-12s %-26s %-35s %s\n", "NODE NAME", "GROUP NAME", "KERNEL VERSION", "VNFS IMAGE", "OVERLAYS (S/R)")
		fmt.Println(strings.Repeat("=", 120))

		for _, node := range nodes {
			fmt.Printf("%-22s %-12s %-26s %-35s %s\n", node.Fqdn.Get(), node.Gid.String(), node.KernelVersion.String(), node.Vnfs.String(), node.SystemOverlay.String()+"/"+node.RuntimeOverlay.String())
		}

	} else {
		fmt.Printf("%-22s %-30s %s\n", "NODE NAME", "VNFS", "PROFILES")
		fmt.Println(strings.Repeat("=", 80))

		for _, node := range nodes {
			fmt.Printf("%-22s %-30s %s\n", node.Fqdn.Get(), node.Vnfs.String(), strings.Join(append(node.GroupProfiles, node.Profiles...), ","))
		}

	}

	return nil
}
