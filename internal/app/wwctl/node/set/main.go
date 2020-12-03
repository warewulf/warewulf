package set

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"os"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var err error
	var nodes []node.NodeInfo

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Cloud not get nodeList: %s\n", err)
		os.Exit(1)
	}

	if SetNodeAll == true {
		nodes, err = nodeDB.FindAllNodes()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Cloud not get nodeList: %s\n", err)
			os.Exit(1)
		}

	} else if len(args) > 0 {
		nodes, err = nodeDB.SearchByNameList(args)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Cloud not get nodeList: %s\n", err)
			os.Exit(1)
		}
	} else {
		cmd.Usage()
		os.Exit(1)
	}

	for _, n := range nodes {
		wwlog.Printf(wwlog.VERBOSE, "Evaluating node: %s\n", n.Id.Get())

		if SetComment != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting comment to: %s\n", n.Id.Get(), SetComment)

			n.Comment.Set(SetComment)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetVnfs != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting vnfs to: %s\n", n.Id.Get(), SetVnfs)

			n.Vnfs.Set(SetVnfs)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetKernel != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting kernel to: %s\n", n.Id.Get(), SetKernel)

			n.KernelVersion.Set(SetKernel)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetDomainName != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting domain name to: %s\n", n.Id.Get(), SetDomainName)

			n.DomainName.Set(SetDomainName)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetIpxe != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting iPXE template to: %s\n", n.Id.Get(), SetIpxe)

			n.Ipxe.Set(SetIpxe)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetRuntimeOverlay != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting runtime overlay to: %s\n", n.Id.Get(), SetRuntimeOverlay)

			n.RuntimeOverlay.Set(SetRuntimeOverlay)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetSystemOverlay != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting system overlay to: %s\n", n.Id.Get(), SetSystemOverlay)

			n.SystemOverlay.Set(SetSystemOverlay)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetIpmiIpaddr != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting IPMI IP address to: %s\n", n.Id.Get(), SetIpmiIpaddr)

			n.IpmiIpaddr.Set(SetIpmiIpaddr)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetIpmiNetmask != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting IPMI netmask to: %s\n", n.Id.Get(), SetIpmiNetmask)

			n.IpmiNetmask.Set(SetIpmiNetmask)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetIpmiUsername != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting IPMI IP username to: %s\n", n.Id.Get(), SetIpmiUsername)

			n.IpmiUserName.Set(SetIpmiUsername)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetIpmiPassword != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting IPMI IP password to: %s\n", n.Id.Get(), SetIpmiPassword)

			n.IpmiPassword.Set(SetIpmiPassword)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}

		if len(SetAddProfile) > 0 {
			for _, p := range SetAddProfile {
				wwlog.Printf(wwlog.VERBOSE, "Node: %s, adding profile to '%s'\n", n.Id.Get(), p)
				n.Profiles = util.SliceAddUniqueElement(n.Profiles, p)
			}
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if len(SetDelProfile) > 0 {
			for _, p := range SetDelProfile {
				wwlog.Printf(wwlog.VERBOSE, "Node: %s, deleting profile from '%s'\n", n.Id.Get(), p)
				n.Profiles = util.SliceRemoveElement(n.Profiles, p)
			}
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}

		if SetNetDevDel == true {
			if SetNetDev == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netdev' option\n")
				os.Exit(1)
			}
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Deleting network device: %s\n", n.Id.Get(), SetNetDev)

			if _, ok := n.NetDevs[SetNetDev]; !ok {
				wwlog.Printf(wwlog.ERROR, "Network Device doesn't exist: %s\n", SetNetDev)
				os.Exit(1)
			}

			delete(n.NetDevs, SetNetDev)
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
				wwlog.Printf(wwlog.ERROR, "Network Device doesn't exist: %s\n", SetNetDev)
				os.Exit(1)
			}
			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting Ipaddr to: %s\n", n.Id.Get(), SetNetDev, SetIpaddr)

			n.NetDevs[SetNetDev].Ipaddr.Set(SetIpaddr)
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
	}

	if len(nodes) > 0 {
		if SetYes == true {
			nodeDB.Persist()
		} else {
			q := fmt.Sprintf("Are you sure you want to modify %d nodes(s)", len(nodes))

			prompt := promptui.Prompt{
				Label:     q,
				IsConfirm: true,
			}

			result, _ := prompt.Run()

			if result == "y" || result == "yes" {
				nodeDB.Persist()
			}
		}

	} else {
		fmt.Printf("No nodes found\n")
	}

	return nil
}
