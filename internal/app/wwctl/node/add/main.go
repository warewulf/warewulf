package add

import (
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	nodeDB, err := node.New()
	if err != nil {
		return errors.Wrap(err, "failed to open node database")
	}

	for _, a := range args {
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

		if SetIpaddr != "" {
			if SetNetDev == "" {
				return errors.New("you must include the '--netdev' option")
			}

			if _, ok := n.NetDevs[SetNetDev]; !ok {
				var netdev node.NetDevEntry
				n.NetDevs[SetNetDev] = &netdev
			}

			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting Ipaddr to: %s\n", n.Id.Get(), SetNetDev, SetIpaddr)

			n.NetDevs[SetNetDev].Ipaddr.Set(SetIpaddr)
			n.NetDevs[SetNetDev].Default.SetB(true)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				return errors.Wrap(err, "failed to update nodedb")
			}
		}
		if SetNetmask != "" {
			if SetNetDev == "" {
				return errors.New("you must include the '--netdev' option")
			}

			if _, ok := n.NetDevs[SetNetDev]; !ok {
				return errors.New("network device does not exist: " + SetNetDev)
			}
			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting netmask to: %s\n", n.Id.Get(), SetNetDev, SetNetmask)

			n.NetDevs[SetNetDev].Netmask.Set(SetNetmask)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				return errors.Wrap(err, "failed to update nodedb")
			}
		}
		if SetGateway != "" {
			if SetNetDev == "" {
				return errors.New("you must include the '--netdev' option")
			}

			if _, ok := n.NetDevs[SetNetDev]; !ok {
				return errors.New("network device does not exist: " + SetNetDev)
			}
			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting gateway to: %s\n", n.Id.Get(), SetNetDev, SetGateway)

			n.NetDevs[SetNetDev].Gateway.Set(SetGateway)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				return errors.Wrap(err, "failed to update nodedb")
			}
		}
		if SetHwaddr != "" {
			if SetNetDev == "" {
				return errors.New("you must include the '--netdev' option")
			}

			if _, ok := n.NetDevs[SetNetDev]; !ok {
				return errors.New("network device does not exist: " + SetNetDev)
			}
			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting HW address to: %s\n", n.Id.Get(), SetNetDev, SetHwaddr)

			n.NetDevs[SetNetDev].Hwaddr.Set(SetHwaddr)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				return errors.Wrap(err, "failed to update nodedb")
			}
		}

		if SetType != "" {
			if SetNetDev == "" {
				return errors.New("you must include the '--netdev' option")
			}

			if _, ok := n.NetDevs[SetNetDev]; !ok {
				return errors.New("network device does not exist: " + SetNetDev)
			}
			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting Type to: %s\n", n.Id.Get(), SetNetDev, SetType)

			n.NetDevs[SetNetDev].Type.Set(SetType)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				return errors.Wrap(err, "failed to update nodedb")
			}
		}

		if SetDiscoverable {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting node to discoverable\n", n.Id.Get())

			n.Discoverable.SetB(true)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				return errors.Wrap(err, "failed to update nodedb")
			}
		}

	}

	return errors.Wrap(nodeDB.Persist(), "failed to persist nodedb")
}
