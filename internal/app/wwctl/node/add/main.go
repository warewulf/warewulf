package add

import (
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/hpcng/warewulf/pkg/hostlist"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var count uint
	nodeDB, err := node.New()
	if err != nil {
		return errors.Wrap(err, "failed to open node database")
	}

	node_args := hostlist.Expand(args)

	for _, a := range node_args {
		n, err := nodeDB.AddNode(a)
		if err != nil {
			return errors.Wrap(err, "failed to add node")
		}
		wwlog.Printf(wwlog.INFO, "Added node: %s\n", a)

		if SetClusterName != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting cluster name to: %s\n", n.Id.Get(), SetClusterName)
			n.ClusterName.Set(SetClusterName)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				return errors.Wrap(err, "failed to update node")
			}
		}

		if SetNetDev != "" {
			if SetNetName == "" {
				return errors.New("you must include the '--netname' option")
			}

			if _, ok := n.NetDevs[SetNetName]; !ok {
				var netdev node.NetDevEntry
				n.NetDevs[SetNetName] = &netdev
			}

			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting Device to: %s\n", n.Id.Get(), SetNetName, SetNetDev)

			n.NetDevs[SetNetName].Device.Set(SetNetDev)
			n.NetDevs[SetNetName].OnBoot.SetB(true)

			//			err := nodeDB.NodeUpdate(n)
			//			if err != nil {
			//				return errors.Wrap(err, "failed to update nodedb")
			//			}
		}

		if SetIpaddr != "" {
			if SetNetName == "" {
				return errors.New("you must include the '--netname' option")
			}

			NewIpaddr := util.IncrementIPv4(SetIpaddr, count)

			if _, ok := n.NetDevs[SetNetName]; !ok {
				var netdev node.NetDevEntry
				n.NetDevs[SetNetName] = &netdev
			}

			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting Ipaddr to: %s\n", n.Id.Get(), SetNetName, NewIpaddr)

			n.NetDevs[SetNetName].Ipaddr.Set(NewIpaddr)
			n.NetDevs[SetNetName].OnBoot.SetB(true)

		}

		if SetNetmask != "" {
			if SetNetName == "" {
				return errors.New("you must include the '--netname' option")
			}

			if _, ok := n.NetDevs[SetNetName]; !ok {
				return errors.New("network device does not exist: " + SetNetName)
			}
			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting netmask to: %s\n", n.Id.Get(), SetNetName, SetNetmask)

			n.NetDevs[SetNetName].Netmask.Set(SetNetmask)
		}

		if SetGateway != "" {
			if SetNetName == "" {
				return errors.New("you must include the '--netname' option")
			}

			if _, ok := n.NetDevs[SetNetName]; !ok {
				return errors.New("network device does not exist: " + SetNetName)
			}
			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting gateway to: %s\n", n.Id.Get(), SetNetName, SetGateway)

			n.NetDevs[SetNetName].Gateway.Set(SetGateway)
		}

		if SetHwaddr != "" {
			if SetNetName == "" {
				return errors.New("you must include the '--netname' option")
			}

			if _, ok := n.NetDevs[SetNetName]; !ok {
				return errors.New("network device does not exist: " + SetNetName)
			}
			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting HW address to: %s\n", n.Id.Get(), SetNetName, SetHwaddr)

			n.NetDevs[SetNetName].Hwaddr.Set(SetHwaddr)
			n.NetDevs[SetNetName].OnBoot.SetB(true)
		}

		if SetType != "" {
			if SetNetName == "" {
				return errors.New("you must include the '--netname' option")
			}

			if _, ok := n.NetDevs[SetNetName]; !ok {
				return errors.New("network device does not exist: " + SetNetName)
			}
			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting Type to: %s\n", n.Id.Get(), SetNetName, SetType)

			n.NetDevs[SetNetName].Type.Set(SetType)
		}

		if SetDiscoverable {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting node to discoverable\n", n.Id.Get())

			n.Discoverable.SetB(true)
		}

		err = nodeDB.NodeUpdate(n)
		if err != nil {
			return errors.Wrap(err, "failed to update nodedb")
		}

		count++
	}

	return errors.Wrap(nodeDB.Persist(), "failed to persist nodedb")
}
