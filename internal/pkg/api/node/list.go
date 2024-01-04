package apinode

import (
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
func NodeList(nodeGet *wwapiv1.GetNodeList) (*wwapiv1.NodeListViewResponse, error) {
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

	var entries []*wwapiv1.NodeListEntry
	if nodeGet.Type == wwapiv1.GetNodeList_Simple {
		for _, n := range node.FilterByName(nodes, nodeGet.Nodes) {
			var netNames []string
			for k := range n.NetDevs {
				netNames = append(netNames, k)
			}
			sort.Strings(netNames)
			entries = append(entries, &wwapiv1.NodeListEntry{
				NodeEntry: &wwapiv1.NodeListEntry_NodeSimple{
					NodeSimple: &wwapiv1.NodeListSimple{
						NodeName: n.Id.Print(),
						Profiles: n.Profiles.Print(),
						Network:  strings.Join(netNames, ", "),
					},
				},
			})
		}
	} else if nodeGet.Type == wwapiv1.GetNodeList_Network {
		for _, n := range node.FilterByName(nodes, nodeGet.Nodes) {
			if len(n.NetDevs) > 0 {
				for name := range n.NetDevs {
					entries = append(entries, &wwapiv1.NodeListEntry{
						NodeEntry: &wwapiv1.NodeListEntry_NodeNetwork{
							NodeNetwork: &wwapiv1.NodeListNetwork{
								NodeName: n.Id.Print(),
								Name:     name,
								Hwaddr:   n.NetDevs[name].Hwaddr.Print(),
								Ipaddr:   n.NetDevs[name].Ipaddr.Print(),
								Gateway:  n.NetDevs[name].Gateway.Print(),
								Device:   n.NetDevs[name].Device.Print(),
							},
						},
					})
				}
			} else {
				entries = append(entries, &wwapiv1.NodeListEntry{
					NodeEntry: &wwapiv1.NodeListEntry_NodeNetwork{
						NodeNetwork: &wwapiv1.NodeListNetwork{
							NodeName: n.Id.Print(),
							Name:     "--",
							Hwaddr:   "--",
							Ipaddr:   "--",
							Gateway:  "--",
							Device:   "--",
						},
					},
				})
			}
		}
	} else if nodeGet.Type == wwapiv1.GetNodeList_Ipmi {
		for _, n := range node.FilterByName(nodes, nodeGet.Nodes) {
			entries = append(entries, &wwapiv1.NodeListEntry{
				NodeEntry: &wwapiv1.NodeListEntry_NodeIpmi{
					NodeIpmi: &wwapiv1.NodeListIpmi{
						NodeName:       n.Id.Print(),
						IpmiIpAddr:     n.Ipmi.Ipaddr.Print(),
						IpmiPort:       n.Ipmi.Port.Print(),
						IpmiUserName:   n.Ipmi.UserName.Print(),
						IpmiInterface:  n.Ipmi.Interface.Print(),
						IpmiEscapeChar: n.Ipmi.EscapeChar.Print(),
					},
				},
			})
		}
	} else if nodeGet.Type == wwapiv1.GetNodeList_Long {
		for _, n := range node.FilterByName(nodes, nodeGet.Nodes) {
			entries = append(entries, &wwapiv1.NodeListEntry{
				NodeEntry: &wwapiv1.NodeListEntry_NodeLong{
					NodeLong: &wwapiv1.NodeListLong{
						NodeName:              n.Id.Print(),
						KernelOverride:        n.Kernel.Override.Print(),
						Container:             n.ContainerName.Print(),
						OverlaysSystemRuntime: n.SystemOverlay.Print() + "/" + n.RuntimeOverlay.Print(),
					},
				},
			})
		}
	} else if nodeGet.Type == wwapiv1.GetNodeList_All || nodeGet.Type == wwapiv1.GetNodeList_FullAll {
		for _, n := range node.FilterByName(nodes, nodeGet.Nodes) {
			fields := n.GetFields(wwapiv1.GetNodeList_FullAll == nodeGet.Type)
			for _, f := range fields {
				entries = append(entries, &wwapiv1.NodeListEntry{
					NodeEntry: &wwapiv1.NodeListEntry_NodeFull{
						NodeFull: &wwapiv1.NodeListFull{
							NodeName: n.Id.Print(),
							Field:    f.Field,
							Profile:  f.Source,
							Value:    f.Value,
						},
					},
				})
			}
		}
	}

	return &wwapiv1.NodeListViewResponse{Nodes: entries}, nil
}
