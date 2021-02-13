package set

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
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

	if SetContainer != "" {
		if container.ValidSource(SetContainer) == true {
			imageFile := container.ImageFile(SetContainer)
			if util.IsFile(imageFile) == false {
				wwlog.Printf(wwlog.ERROR, "Container has not been built: %s\n", SetContainer)
				if SetForce == false {
					os.Exit(1)
				}
			}
		} else {
			wwlog.Printf(wwlog.ERROR, "Container does not exist: %s\n", SetContainer)
			if SetForce == false {
				os.Exit(1)
			}
		}
	}

	for _, n := range nodes {
		wwlog.Printf(wwlog.VERBOSE, "Evaluating node: %s\n", n.Id.Get())

		if SetComment != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting comment to: %s\n", n.Id.Get(), SetComment)
			n.Comment.Set(SetComment)
		}

		if SetContainer != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting container name to: %s\n", n.Id.Get(), SetContainer)
			n.ContainerName.Set(SetContainer)
		}

		if SetInit != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting init command to: %s\n", n.Id.Get(), SetInit)
			n.Init.Set(SetInit)
		}

		if SetRoot != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting root to: %s\n", n.Id.Get(), SetRoot)
			n.Root.Set(SetRoot)
		}

		if SetKernel != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting kernel to: %s\n", n.Id.Get(), SetKernel)
			n.KernelVersion.Set(SetKernel)
		}

		if SetKernelArgs != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting kernel args to: %s\n", n.Id.Get(), SetKernelArgs)
			n.KernelArgs.Set(SetKernelArgs)
		}

		if SetClusterName != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting cluster name to: %s\n", n.Id.Get(), SetClusterName)
			n.ClusterName.Set(SetClusterName)
		}

		if SetIpxe != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting iPXE template to: %s\n", n.Id.Get(), SetIpxe)
			n.Ipxe.Set(SetIpxe)
		}

		if SetRuntimeOverlay != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting runtime overlay to: %s\n", n.Id.Get(), SetRuntimeOverlay)
			n.RuntimeOverlay.Set(SetRuntimeOverlay)
		}

		if SetSystemOverlay != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting system overlay to: %s\n", n.Id.Get(), SetSystemOverlay)
			n.SystemOverlay.Set(SetSystemOverlay)
		}

		if SetIpmiIpaddr != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting IPMI IP address to: %s\n", n.Id.Get(), SetIpmiIpaddr)
			n.IpmiIpaddr.Set(SetIpmiIpaddr)
		}

		if SetIpmiNetmask != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting IPMI netmask to: %s\n", n.Id.Get(), SetIpmiNetmask)
			n.IpmiNetmask.Set(SetIpmiNetmask)
		}

		if SetIpmiGateway != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting IPMI gateway to: %s\n", n.Id.Get(), SetIpmiGateway)
			n.IpmiGateway.Set(SetIpmiGateway)
		}

		if SetIpmiUsername != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting IPMI IP username to: %s\n", n.Id.Get(), SetIpmiUsername)
			n.IpmiUserName.Set(SetIpmiUsername)
		}

		if SetIpmiPassword != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting IPMI IP password to: %s\n", n.Id.Get(), SetIpmiPassword)
			n.IpmiPassword.Set(SetIpmiPassword)
		}

		if SetDiscoverable == true {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting node to discoverable\n", n.Id.Get())
			n.Discoverable.SetB(true)
		}

		if SetUndiscoverable == true {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting node to undiscoverable\n", n.Id.Get())
			n.Discoverable.SetB(false)
		}

		if len(SetAddProfile) > 0 {
			for _, p := range SetAddProfile {
				wwlog.Printf(wwlog.VERBOSE, "Node: %s, adding profile to '%s'\n", n.Id.Get(), p)
				n.Profiles = util.SliceAddUniqueElement(n.Profiles, p)
			}
		}

		if len(SetDelProfile) > 0 {
			for _, p := range SetDelProfile {
				wwlog.Printf(wwlog.VERBOSE, "Node: %s, deleting profile from '%s'\n", n.Id.Get(), p)
				n.Profiles = util.SliceRemoveElement(n.Profiles, p)
			}
		}

		if SetNetDevDel == true {
			if SetNetDev == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netdev' option\n")
				os.Exit(1)
			}

			if _, ok := n.NetDevs[SetNetDev]; !ok {
				wwlog.Printf(wwlog.ERROR, "Network Device doesn't exist: %s\n", SetNetDev)
				os.Exit(1)
			}

			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Deleting network device: %s\n", n.Id.Get(), SetNetDev)
			delete(n.NetDevs, SetNetDev)
		}
		if SetIpaddr != "" {
			if SetNetDev == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netdev' option\n")
				os.Exit(1)
			}

			if _, ok := n.NetDevs[SetNetDev]; !ok {
				var nd node.NetDevEntry
				n.NetDevs[SetNetDev] = &nd
			}

			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting Ipaddr to: %s\n", n.Id.Get(), SetNetDev, SetIpaddr)
			n.NetDevs[SetNetDev].Ipaddr.Set(SetIpaddr)
		}
		if SetNetmask != "" {
			if SetNetDev == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netdev' option\n")
				os.Exit(1)
			}

			if _, ok := n.NetDevs[SetNetDev]; !ok {
				var nd node.NetDevEntry
				n.NetDevs[SetNetDev] = &nd
			}

			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting netmask to: %s\n", n.Id.Get(), SetNetDev, SetNetmask)
			n.NetDevs[SetNetDev].Netmask.Set(SetNetmask)
		}
		if SetGateway != "" {
			if SetNetDev == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netdev' option\n")
				os.Exit(1)
			}

			if _, ok := n.NetDevs[SetNetDev]; !ok {
				var nd node.NetDevEntry
				n.NetDevs[SetNetDev] = &nd
			}

			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting gateway to: %s\n", n.Id.Get(), SetNetDev, SetGateway)
			n.NetDevs[SetNetDev].Gateway.Set(SetGateway)
		}
		if SetHwaddr != "" {
			if SetNetDev == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netdev' option\n")
				os.Exit(1)
			}

			if _, ok := n.NetDevs[SetNetDev]; !ok {
				var nd node.NetDevEntry
				n.NetDevs[SetNetDev] = &nd
			}

			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting HW address to: %s\n", n.Id.Get(), SetNetDev, SetHwaddr)
			n.NetDevs[SetNetDev].Hwaddr.Set(SetHwaddr)
		}
		if SetNetDevDefault == true {
			if SetNetDev == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netdev' option\n")
				os.Exit(1)
			}

			if _, ok := n.NetDevs[SetNetDev]; !ok {
				var nd node.NetDevEntry
				n.NetDevs[SetNetDev] = &nd
			}

			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting device as default\n", n.Id.Get(), SetNetDev)
			for _, dev := range n.NetDevs {
				// First clear all other devices that might be configured as default
				dev.Default.SetB(false)
			}
			n.NetDevs[SetNetDev].Default.SetB(true)
		}

		err := nodeDB.NodeUpdate(n)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
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
				warewulfd.DaemonReload()
			}
		}

	} else {
		fmt.Printf("No nodes found\n")
	}

	return nil
}
