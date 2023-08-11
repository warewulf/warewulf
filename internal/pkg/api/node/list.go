package apinode

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/pkg/hostlist"
)

/*
NodeList lists all to none of the nodes managed by Warewulf. Returns
a formated string slice, with each line as separate string
*/
func NodeList(nodeGet *wwapiv1.GetNodeList) (nodeList wwapiv1.NodeList, err error) {
	// nil is okay for nodeNames
	nodeDB, err := node.New()
	if err != nil {
		return
	}
	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		return
	}
	nodeGet.Nodes = hostlist.Expand(nodeGet.Nodes)
	sort.Strings(nodeGet.Nodes)
	if nodeGet.Type == wwapiv1.GetNodeList_Simple {
		nodeList.Output = append(nodeList.Output,
			fmt.Sprintf("%-22s %-26s %s", "NODE NAME", "PROFILES", "NETWORK"))
		nodeList.Output = append(nodeList.Output, (strings.Repeat("=", 80)))
		for _, n := range node.FilterByName(nodes, nodeGet.Nodes) {
			var netNames []string
			for k := range n.NetDevs {
				netNames = append(netNames, k)
			}
			nodeList.Output = append(nodeList.Output,
				fmt.Sprintf("%-22s %-26s %s", n.Id.Print(), n.Profiles.Print(), strings.Join(netNames, ", ")))
		}
	} else if nodeGet.Type == wwapiv1.GetNodeList_Network {
		nodeList.Output = append(nodeList.Output,
			fmt.Sprintf("%-22s %-8s %-18s %-15s %-15s %-15s", "NODE NAME", "NAME", "HWADDR", "IPADDR", "GATEWAY", "DEVICE"),
			strings.Repeat("=", 90))
		for _, n := range node.FilterByName(nodes, nodeGet.Nodes) {
			if len(n.NetDevs) > 0 {
				for name := range n.NetDevs {
					nodeList.Output = append(nodeList.Output,
						fmt.Sprintf("%-22s %-8s %-18s %-15s %-15s %-15s", n.Id.Print(), name,
							n.NetDevs[name].Hwaddr.Print(),
							n.NetDevs[name].Ipaddr.Print(),
							n.NetDevs[name].Gateway.Print(),
							n.NetDevs[name].Device.Print()))
				}
			} else {
				fmt.Printf("%-22s %-6s %-18s %-15s %-15s", n.Id.Print(), "--", "--", "--", "--")
			}
		}
	} else if nodeGet.Type == wwapiv1.GetNodeList_Ipmi {
		nodeList.Output = append(nodeList.Output,
			fmt.Sprintf("%-22s %-16s %-10s %-20s %-14s", "NODE NAME", "IPMI IPADDR", "IPMI PORT", "IPMI USERNAME", "IPMI INTERFACE"),
			strings.Repeat("=", 98))
		for _, n := range node.FilterByName(nodes, nodeGet.Nodes) {
			nodeList.Output = append(nodeList.Output,
				fmt.Sprintf("%-22s %-16s %-10s %-20s %-14s", n.Id.Print(),
					n.Ipmi.Ipaddr.Print(),
					n.Ipmi.Port.Print(),
					n.Ipmi.UserName.Print(),
					n.Ipmi.Interface.Print()))
		}
	} else if nodeGet.Type == wwapiv1.GetNodeList_Long {
		nodeList.Output = append(nodeList.Output,
			fmt.Sprintf("%-22s %-16s %-16s %s", "NODE NAME", "KERNEL OVERRIDE", "CONTAINER", "OVERLAYS (S/R)"),
			strings.Repeat("=", 85))
		for _, n := range node.FilterByName(nodes, nodeGet.Nodes) {
			nodeList.Output = append(nodeList.Output,
				fmt.Sprintf("%-22s %-16s %-16s %s", n.Id.Print(),
					n.Kernel.Override.Print(),
					n.ContainerName.Print(),
					n.SystemOverlay.Print()+"/"+n.RuntimeOverlay.Print()))
		}
	} else if nodeGet.Type == wwapiv1.GetNodeList_All || nodeGet.Type == wwapiv1.GetNodeList_FullAll {
		for _, n := range node.FilterByName(nodes, nodeGet.Nodes) {
			nodeList.Output = append(nodeList.Output,
				fmt.Sprintf("%s:=:%s:=:%s:=:%s", "NODE", "FIELD", "PROFILE", "VALUE"))
			fields := n.GetFields(wwapiv1.GetNodeList_FullAll == nodeGet.Type)
			for _, f := range fields {
				nodeList.Output = append(nodeList.Output,
					fmt.Sprintf("%s:=:%s:=:%s:=:%s", n.Id.Print(), f.Field, f.Source, f.Value))
			}
		}
	}
	return
}
