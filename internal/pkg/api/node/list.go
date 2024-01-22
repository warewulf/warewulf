package apinode

import (
	"sort"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/pkg/hostlist"
)

func NodeList(nodeGet *wwapiv1.GetNodeList) (*node.NodeListResponse, error) {
	// nil is okay for nodeNames
	nodeDB, err := node.New()
	if err != nil {
		return nil, err
	}
	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		return nil, err
	}
	nodeGet.Nodes = hostlist.Expand(nodeGet.Nodes)
	sort.Strings(nodeGet.Nodes)

	resp := &node.NodeListResponse{
		Nodes: make(map[string][]node.NodeListEntry),
	}

	if nodeGet.Type == wwapiv1.GetNodeList_Simple {
		for _, n := range node.FilterByName(nodes, nodeGet.Nodes) {
			var netNames []string
			for k := range n.NetDevs {
				netNames = append(netNames, k)
			}
			sort.Strings(netNames)

			var entries []node.NodeListEntry
			entries = append(entries, &node.NodeListSimpleEntry{
				Profile: n.Profiles.Print(),
				Network: strings.Join(netNames, ", "),
			})

			if vals, ok := resp.Nodes[n.Id.Print()]; ok {
				entries = append(entries, vals...)
			}
			resp.Nodes[n.Id.Print()] = entries
		}
	} else if nodeGet.Type == wwapiv1.GetNodeList_Network {
		for _, n := range node.FilterByName(nodes, nodeGet.Nodes) {
			var entries []node.NodeListEntry
			if len(n.NetDevs) > 0 {
				for name := range n.NetDevs {
					entries = append(entries, &node.NodeListNetworkEntry{
						Name:    name,
						HwAddr:  n.NetDevs[name].Hwaddr.Print(),
						IpAddr:  n.NetDevs[name].Ipaddr.Print(),
						Gateway: n.NetDevs[name].Gateway.Print(),
						Device:  n.NetDevs[name].Device.Print(),
					})
				}
			} else {
				entries = append(entries, &node.NodeListNetworkEntry{
					Name:    "--",
					HwAddr:  "--",
					IpAddr:  "--",
					Gateway: "--",
					Device:  "--",
				})
			}

			if vals, ok := resp.Nodes[n.Id.Print()]; ok {
				entries = append(entries, vals...)
			}
			resp.Nodes[n.Id.Print()] = entries
		}
	} else if nodeGet.Type == wwapiv1.GetNodeList_Ipmi {
		for _, n := range node.FilterByName(nodes, nodeGet.Nodes) {
			var entries []node.NodeListEntry
			entries = append(entries, &node.NodeListIpmiEntry{
				IpmiAddr:       n.Ipmi.Ipaddr.Print(),
				IpmiPort:       n.Ipmi.Port.Print(),
				IpmiUser:       n.Ipmi.UserName.Print(),
				IpmiInterface:  n.Ipmi.Interface.Print(),
				IpmiEscapeChar: n.Ipmi.EscapeChar.Print(),
			})

			if vals, ok := resp.Nodes[n.Id.Print()]; ok {
				entries = append(entries, vals...)
			}
			resp.Nodes[n.Id.Print()] = entries
		}
	} else if nodeGet.Type == wwapiv1.GetNodeList_Long {
		for _, n := range node.FilterByName(nodes, nodeGet.Nodes) {
			var entries []node.NodeListEntry
			entries = append(entries, &node.NodeListLongEntry{
				KernelOverride: n.Kernel.Override.Print(),
				Container:      n.ContainerName.Print(),
				Overlays:       n.SystemOverlay.Print() + "/" + n.RuntimeOverlay.Print(),
			})

			if vals, ok := resp.Nodes[n.Id.Print()]; ok {
				entries = append(entries, vals...)
			}
			resp.Nodes[n.Id.Print()] = entries
		}
	} else if nodeGet.Type == wwapiv1.GetNodeList_All || nodeGet.Type == wwapiv1.GetNodeList_FullAll {
		for _, n := range node.FilterByName(nodes, nodeGet.Nodes) {
			var entries []node.NodeListEntry
			fields := n.GetFields(wwapiv1.GetNodeList_FullAll == nodeGet.Type)
			for _, f := range fields {
				entries = append(entries, &node.NodeListAllEntry{
					Field:   f.Field,
					Profile: f.Source,
					Value:   f.Value,
				})
			}

			if vals, ok := resp.Nodes[n.Id.Print()]; ok {
				entries = append(entries, vals...)
			}
			resp.Nodes[n.Id.Print()] = entries
		}
	}

	return resp, nil
}
