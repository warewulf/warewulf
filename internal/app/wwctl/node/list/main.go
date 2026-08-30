package list

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/warewulf/warewulf/internal/app/wwctl/table"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		nodeDB, err := node.New()
		if err != nil {
			return
		}
		nodes, err := nodeDB.FindAllNodes()
		if err != nil {
			return
		}
		nodeNames := hostlist.Expand(args)
		sort.Strings(nodeNames)
		filtered := node.FilterNodeListByName(nodes, nodeNames)

		if vars.showYaml || vars.showJson {
			nodeMap := make(map[string]node.Node)
			for _, n := range filtered {
				nodeMap[n.Id()] = n
			}
			var buf []byte
			if vars.showJson {
				buf, _ = json.MarshalIndent(nodeMap, "", "  ")
			} else {
				buf, _ = yaml.Marshal(nodeMap)
			}
			wwlog.Info(string(buf))
		} else if vars.showAll {
			t := table.New(cmd.OutOrStdout())
			t.AddHeader("NODE", "FIELD", "PROFILE", "VALUE")
			for _, n := range filtered {
				if _, fields, err := nodeDB.MergeNode(n.Id()); err != nil {
					wwlog.Error("unable to merge node %v: %v", n.Id(), err)
					continue
				} else {
					for _, f := range fields.List(n) {
						t.AddLine(table.Prep([]string{n.Id(), f.Field, f.Source, f.Value})...)
					}
				}
			}
			t.Print()
		} else if vars.showIpmi {
			t := table.New(cmd.OutOrStdout())
			t.AddHeader("NODE", "IPMI IPADDR", "IPMI PORT", "IPMI USERNAME", "IPMI INTERFACE")
			for _, n := range filtered {
				ipaddr, port, username, iface := "", "", "", ""
				if n.Ipmi != nil {
					ipaddr = n.Ipmi.Ipaddr.String()
					port = n.Ipmi.Port
					username = n.Ipmi.UserName
					iface = n.Ipmi.Interface
				}
				t.AddLine(table.Prep([]string{n.Id(), ipaddr, port, username, iface})...)
			}
			t.Print()
		} else if vars.showNet {
			t := table.New(cmd.OutOrStdout())
			t.AddHeader("NODE", "NETWORK", "HWADDR", "IPADDR", "GATEWAY", "DEVICE")
			for _, n := range filtered {
				if len(n.NetDevs) > 0 {
					for name := range n.NetDevs {
						t.AddLine(table.Prep([]string{n.Id(), name,
							n.NetDevs[name].Hwaddr,
							n.NetDevs[name].Ipaddr.String(),
							n.NetDevs[name].Gateway.String(),
							n.NetDevs[name].Device})...)
					}
				} else {
					t.AddLine(table.Prep([]string{n.Id(), "", "", "", "", ""})...)
				}
			}
			t.Print()
		} else if vars.showLong {
			t := table.New(cmd.OutOrStdout())
			t.AddHeader("NODE NAME", "KERNEL VERSION", "IMAGE", "OVERLAYS (S/R)")
			for _, n := range filtered {
				kernelVersion := ""
				if n.Kernel != nil {
					kernelVersion = n.Kernel.Version
				}
				t.AddLine(table.Prep([]string{n.Id(),
					kernelVersion,
					n.ImageName,
					strings.Join(n.SystemOverlay, ",") + "/" + strings.Join(n.RuntimeOverlay, ",")})...)
			}
			t.Print()
		} else {
			// Simple (default)
			t := table.New(cmd.OutOrStdout())
			t.AddHeader("NODE NAME", "PROFILES", "NETWORK")
			for _, n := range filtered {
				var netNames []string
				for k := range n.NetDevs {
					netNames = append(netNames, k)
				}
				sort.Strings(netNames)
				t.AddLine(table.Prep([]string{n.Id(), strings.Join(n.Profiles, ","), strings.Join(netNames, ", ")})...)
			}
			t.Print()
		}
		return
	}
}
