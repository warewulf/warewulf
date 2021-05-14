package add

import (
	"fmt"
	"os"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Failed to open node database: %s\n", err)
		os.Exit(1)
	}

	for _, a := range args {
		n, err := nodeDB.AddNode(a)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Added node: %s\n", a)

	        if SetClusterName != "" {
                        wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting cluster name to: %s\n", n.Id.Get(), SetClusterName)
                        n.ClusterName.Set(SetClusterName)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
                                wwlog.Printf(wwlog.ERROR, "%s\n", err)
                                os.Exit(1)
                        }
                }

		if SetIpaddr != "" {
			if SetNetDev == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netdev' option\n")
				os.Exit(1)
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
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetNetmask != "" {
			if SetNetDev == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netdev' option\n")
				os.Exit(1)
			}

			if _, ok := n.NetDevs[SetNetDev]; !ok {
				wwlog.Printf(wwlog.ERROR, "Network Device doesn't exist: %s\n", SetNetDev)
				os.Exit(1)
			}
			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting netmask to: %s\n", n.Id.Get(), SetNetDev, SetNetmask)

			n.NetDevs[SetNetDev].Netmask.Set(SetNetmask)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetGateway != "" {
			if SetNetDev == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netdev' option\n")
				os.Exit(1)
			}

			if _, ok := n.NetDevs[SetNetDev]; !ok {
				wwlog.Printf(wwlog.ERROR, "Network Device doesn't exist: %s\n", SetNetDev)
				os.Exit(1)
			}
			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting gateway to: %s\n", n.Id.Get(), SetNetDev, SetGateway)

			n.NetDevs[SetNetDev].Gateway.Set(SetGateway)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetHwaddr != "" {
			if SetNetDev == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netdev' option\n")
				os.Exit(1)
			}

			if _, ok := n.NetDevs[SetNetDev]; !ok {
				wwlog.Printf(wwlog.ERROR, "Network Device doesn't exist: %s\n", SetNetDev)
				os.Exit(1)
			}
			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting HW address to: %s\n", n.Id.Get(), SetNetDev, SetHwaddr)

			n.NetDevs[SetNetDev].Hwaddr.Set(SetHwaddr)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}

		if SetType != "" {
			if SetNetDev == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netdev' option\n")
				os.Exit(1)
			}

			if _, ok := n.NetDevs[SetNetDev]; !ok {
				wwlog.Printf(wwlog.ERROR, "Network Device doesn't exist: %s\n", SetNetDev)
				os.Exit(1)
			}
			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting Type to: %s\n", n.Id.Get(), SetNetDev, SetType)

			n.NetDevs[SetNetDev].Type.Set(SetType)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}

		if SetDiscoverable == true {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting node to discoverable\n", n.Id.Get())

			n.Discoverable.SetB(true)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}

	}

	nodeDB.Persist()

	return nil
}
