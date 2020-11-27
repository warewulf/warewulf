package list

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
	"reflect"
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
			v := reflect.ValueOf(node)
			typeOfS := v.Type()
			fmt.Printf("################################################################################\n")
			for i := 0; i< v.NumField(); i++ {
				//TODO: Fix for NetDevs and Interface should print Fprint() method
				fmt.Printf("%-25s %s = %#v\n", node.Fqdn.Get(), typeOfS.Field(i).Name, v.Field(i).Interface())
			}
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
			fmt.Printf("%-22s %-16s %-20s %-20s\n", node.Fqdn.Get(), node.IpmiIpaddr.Get(), node.IpmiUserName.Get(), node.IpmiPassword.Get())
		}

	} else if ShowLong == true {
		fmt.Printf("%-22s %-12s %-26s %-35s %s\n", "NODE NAME", "GROUP NAME", "KERNEL VERSION", "VNFS IMAGE", "OVERLAYS (S/R)")
		fmt.Println(strings.Repeat("=", 120))

		for _, node := range nodes {
			fmt.Printf("%-22s %-12s %-26s %-35s %s\n", node.Fqdn.Get(), node.GroupName.Get(), node.KernelVersion.Get(), node.Vnfs.Get(), node.SystemOverlay.Get() +"/"+ node.RuntimeOverlay.Get())
		}

	} else {
		fmt.Printf("%-22s %s\n", "NODE NAME", "PROFILES")
		fmt.Println(strings.Repeat("=", 80))

		for _, node := range nodes {
			fmt.Printf("%-22s %s\n", node.Fqdn.Get(), append(node.GroupProfiles, node.Profiles...))
		}

	}


	return nil
}
