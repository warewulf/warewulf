package list

import (
	"fmt"
	"sort"
	"strings"

	apinode "github.com/hpcng/warewulf/internal/pkg/api/node"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {

	nodeInfo, err := apinode.NodeList(args)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		return
	}
	nodeLopt := node.GetloptMap(node.NodeConf{})
	ipmiLopt := node.GetloptMap(node.IpmiConf{})
	kernelLopt := node.GetloptMap(node.KernelConf{})
	netdevLopt := node.GetloptMap(node.NetDevs{})
	if ShowAll {
		for i := 0; i < len(nodeInfo); i++ {
			ni := nodeInfo[i]
			nodeName := `UNKNOWN`
			if _, ok := ni.Fields["Id"]; ok {
				nodeName = ni.Fields["Id"].Print
			}
			var nodeconfkeys, ipmiconfkeys, kernelconfkeys, netdevkys []string
			for k := range ni.Fields {
				subkeys := strings.Split(k, ":")
				if len(subkeys) == 1 {
					nodeconfkeys = append(nodeconfkeys, k)
				}
				if len(subkeys) >= 2 {
					switch subkeys[0] {
					case "IpmiEntry":
						ipmiconfkeys = append(ipmiconfkeys, k)
					case "KernelEntry":
						kernelconfkeys = append(kernelconfkeys, k)
					case "NetDevEntry":
						netdevkys = append(netdevkys, k)
					}
				}
			}
			sort.Strings(nodeconfkeys)
			sort.Strings(ipmiconfkeys)
			sort.Strings(kernelconfkeys)
			sort.Strings(netdevkys)
			keyssorted := append(nodeconfkeys, kernelconfkeys...)
			keyssorted = append(keyssorted, ipmiconfkeys...)
			keyssorted = append(keyssorted, netdevkys...)
			fmt.Printf("################################################################################\n")
			fmt.Printf("%-20s %-18s %-12s %s\n", "NODE", "FIELD", "PROFILE", "VALUE")
			for _, keys := range keyssorted {
				fieldName := keys
				subkeys := strings.Split(keys, ":")
				if len(subkeys) == 1 {
					if subkeys[0] == "Id" {
						continue
					}
					fieldName = nodeLopt[subkeys[0]]
				}
				if len(subkeys) >= 2 {
					switch subkeys[0] {
					case "IpmiEntry":
						fieldName = ipmiLopt[subkeys[1]]
					case "KernelEntry":
						fieldName = kernelLopt[subkeys[1]]
					case "NetDevEntry":
						if len(subkeys) == 3 {
							fieldName = subkeys[1] + ":" + netdevLopt[subkeys[2]]
						} else if len(subkeys) == 4 {
							fieldName = subkeys[1] + ":keys:" + subkeys[3]
						}
					}
				}
				fmt.Printf("%-20s %-18s %-12s %s\n", nodeName, fieldName, ni.Fields[keys].Source, ni.Fields[keys].Print)
			}
		}
	} else if ShowNet {
		fmt.Printf("%-22s %-8s %-18s %-15s %-15s %-15s\n", "NODE NAME", "NAME", "HWADDR", "IPADDR", "GATEWAY", "DEVICE")
		fmt.Println(strings.Repeat("=", 90))

		for i := 0; i < len(nodeInfo); i++ {
			ni := nodeInfo[i]
			nodeName := `UNKNOWN`
			if _, ok := ni.Fields["Id"]; ok {
				nodeName = ni.Fields["Id"].Print
			}
			netNames := make(map[string]bool)
			for k := range ni.Fields {
				subkeys := strings.Split(k, ":")
				if len(subkeys) == 3 && subkeys[0] == "NetDevEntry" {
					netNames[subkeys[1]] = true
				}
			}
			if len(netNames) > 0 {
				for name := range netNames {
					fmt.Printf("%-22s %-8s %-18s %-15s %-15s %-15s\n", nodeName, name,
						ni.Fields["NetDevEntry:"+name+":Hwaddr"].Print,
						ni.Fields["NetDevEntry:"+name+":Ipaddr"].Print,
						ni.Fields["NetDevEntry:"+name+":Gateway"].Print,
						ni.Fields["NetDevEntry:"+name+":Device"].Print)
				}
			} else {
				fmt.Printf("%-22s %-6s %-18s %-15s %-15s\n", nodeName, "--", "--", "--", "--")
			}
		}
	} else if ShowIpmi {
		fmt.Printf("%-22s %-16s %-10s %-20s %-14s\n", "NODE NAME", "IPMI IPADDR", "IPMI PORT", "IPMI USERNAME", "IPMI INTERFACE")
		fmt.Println(strings.Repeat("=", 98))

		for i := 0; i < len(nodeInfo); i++ {
			ni := nodeInfo[i]
			nodeName := `UNKNOWN`
			if _, ok := ni.Fields["Id"]; ok {
				nodeName = ni.Fields["Id"].Print
			}
			fmt.Printf("%-22s %-16s %-10s %-20s %-14s\n", nodeName,
				ni.Fields["IpmiEntry:Ipaddr"].Print,
				ni.Fields["IpmiEntry:Port"].Print,
				ni.Fields["IpmiEntry:UserName"].Print,
				ni.Fields["IpmiEntry:Interface"].Print)
		}

	} else if ShowLong {
		fmt.Printf("%-22s %-16s %-16s %s\n", "NODE NAME", "KERNEL OVERRIDE", "CONTAINER", "OVERLAYS (S/R)")
		fmt.Println(strings.Repeat("=", 85))

		for i := 0; i < len(nodeInfo); i++ {
			ni := nodeInfo[i]
			nodeName := `UNKNOWN`
			if _, ok := ni.Fields["Id"]; ok {
				nodeName = ni.Fields["Id"].Print
			}
			fmt.Printf("%-22s %-16s %-16s %s\n", nodeName,
				ni.Fields["KernelEntry:Override"].Print,
				ni.Fields["ContainerName"].Print,
				ni.Fields["SystemOverlay"].Print+"/"+ni.Fields["RuntimeOverlay"].Print)
		}

	} else {
		fmt.Printf("%-22s %-26s %s\n", "NODE NAME", "PROFILES", "NETWORK")
		fmt.Println(strings.Repeat("=", 80))
		for i := 0; i < len(nodeInfo); i++ {
			ni := nodeInfo[i]
			nodeName := `UNKNOWN`
			if _, ok := ni.Fields["Id"]; ok {
				nodeName = ni.Fields["Id"].Print
			}
			netNameMap := make(map[string]bool)
			var netNames []string
			for k := range ni.Fields {
				subkeys := strings.Split(k, ":")
				if len(subkeys) == 3 && subkeys[0] == "NetDevEntry" && !netNameMap[subkeys[1]] {
					netNameMap[subkeys[1]] = true
					netNames = append(netNames, subkeys[1])
				}
			}
			fmt.Printf("%-22s %-26s %s\n", nodeName, ni.Fields["Profiles"].Print, strings.Join(netNames, ", "))
		}
	}
	return
}
