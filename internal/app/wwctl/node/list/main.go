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
				fmt.Printf("%-25s %s = %v\n", node.Fqdn.String(), typeOfS.Field(i).Name, v.Field(i).Interface())
			}
		}

	} else if ShowNet == true {
		fmt.Printf("%-22s %-6s %-18s %-15s %-15s\n", "NODE NAME", "DEVICE", "HWADDR", "IPADDR", "GATEWAY")
		fmt.Println(strings.Repeat("=", 80))

		for _, node := range nodes {
			if len(node.NetDevs) > 0 {
				for name, dev := range node.NetDevs {
					fmt.Printf("%-22s %-6s %-18s %-15s %-15s\n", node.Fqdn.String(), name, dev.Hwaddr, dev.Ipaddr, dev.Gateway)
				}
			} else {
				fmt.Printf("%-22s %-6s %-18s %-15s %-15s\n", node.Fqdn.String(), "--", "--", "--", "--")
			}
		}

	} else if ShowIpmi == true {
		fmt.Printf("%-22s %-16s %-20s %-20s\n", "NODE NAME", "IPMI IPADDR", "IPMI USERNAME", "IPMI PASSWORD")
		fmt.Println(strings.Repeat("=", 80))

		for _, node := range nodes {
			fmt.Printf("%-22s %-16s %-20s %-20s\n", node.Fqdn.String(), node.IpmiIpaddr.Fprint(), node.IpmiUserName.Fprint(), node.IpmiPassword.Fprint())
		}

	} else if ShowLong == true {
		fmt.Printf("%-22s %-12s %-26s %-30s %-12s\n", "NODE NAME", "GROUP NAME", "KERNEL VERSION", "VNFS IMAGE", "R-OVERLAY")
		fmt.Println(strings.Repeat("=", 100))

		for _, node := range nodes {
			fmt.Printf("%-22s %-12s %-26s %-30s %-12s\n", node.Fqdn.String(), node.GroupName.Fprint(), node.KernelVersion.Fprint(), node.Vnfs.Fprint(), node.RuntimeOverlay.Fprint())
		}

	} else {
		fmt.Printf("%-22s %-26s %-30s\n", "NODE NAME", "KERNEL VERSION", "VNFS IMAGE")
		fmt.Println(strings.Repeat("=", 80))

		for _, node := range nodes {
			fmt.Printf("%-22s %-26s %-30s\n", node.Fqdn.String(), node.KernelVersion.Fprint(), node.Vnfs.Fprint())
		}

	}


	return nil
}
