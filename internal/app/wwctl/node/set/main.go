package set

import (
	"fmt"
	"os"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var err error
	var count uint
	var SetProfiles []string

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not get node list: %s\n", err)
		os.Exit(1)
	}

	if SetNodeAll || (len(args) == 0 && len(nodes) > 0) {
		fmt.Printf("\n*** WARNING: This command will modify all nodes! ***\n\n")
	} else {
		nodes = node.FilterByName(nodes, args)
	}

	if len(nodes) == 0 {
		fmt.Printf("No nodes found\n")
		os.Exit(1)
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

		if SetAssetKey != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting asset key to: %s\n", n.Id.Get(), SetAssetKey)
			n.AssetKey.Set(SetAssetKey)
		}

		if SetKernelOverride != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting kernel override to: %s\n", n.Id.Get(), SetKernelOverride)
			n.KernelOverride.Set(SetKernelOverride)
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

		if len(SetRuntimeOverlay) != 0 {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting runtime overlay to: %s\n", n.Id.Get(), SetRuntimeOverlay)
			n.RuntimeOverlay.SetSlice(SetRuntimeOverlay)
		}

		if len(SetSystemOverlay) != 0 {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting system overlay to: %s\n", n.Id.Get(), SetSystemOverlay)
			n.SystemOverlay.SetSlice(SetSystemOverlay)
		}

		if SetIpmiIpaddr != "" {
			NewIpaddr := util.IncrementIPv4(SetIpmiIpaddr, count)
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting IPMI IP address to: %s\n", n.Id.Get(), NewIpaddr)
			n.IpmiIpaddr.Set(NewIpaddr)
		}

		if SetIpmiNetmask != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting IPMI netmask to: %s\n", n.Id.Get(), SetIpmiNetmask)
			n.IpmiNetmask.Set(SetIpmiNetmask)
		}

		if SetIpmiPort != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting IPMI port to: %s\n", n.Id.Get(), SetIpmiPort)
			n.IpmiPort.Set(SetIpmiPort)
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

		if SetIpmiInterface != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting IPMI IP interface to: %s\n", n.Id.Get(), SetIpmiInterface)
			n.IpmiInterface.Set(SetIpmiInterface)
		}

		if SetIpmiWrite == "yes" || SetNetOnBoot == "y" || SetNetOnBoot == "1" || SetNetOnBoot == "true" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting Ipmiwrite to %s\n", n.Id.Get(), SetIpmiWrite)
			n.IpmiWrite.SetB(true)
		} else {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting Ipmiwrite to %s\n", n.Id.Get(), SetIpmiWrite)
			n.IpmiWrite.SetB(false)
		}

		if SetDiscoverable {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting node to discoverable\n", n.Id.Get())
			n.Discoverable.SetB(true)
		}

		if SetUndiscoverable {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting node to undiscoverable\n", n.Id.Get())
			n.Discoverable.SetB(false)
		}

		if len(SetProfiles) > 0 {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting profiles to: %s\n", n.Id.Get(), strings.Join(SetProfiles, ","))
			n.Profiles = SetProfiles
		}

		if len(SetAddProfile) > 0 {
			for _, p := range SetAddProfile {
				wwlog.Printf(wwlog.VERBOSE, "Node: %s, adding profile '%s'\n", n.Id.Get(), p)
				n.Profiles = util.SliceAddUniqueElement(n.Profiles, p)
			}
		}

		if len(SetDelProfile) > 0 {
			for _, p := range SetDelProfile {
				wwlog.Printf(wwlog.VERBOSE, "Node: %s, deleting profile '%s'\n", n.Id.Get(), p)
				n.Profiles = util.SliceRemoveElement(n.Profiles, p)
			}
		}

		if SetNetName != "" {
			if _, ok := n.NetDevs[SetNetName]; !ok {
				var nd node.NetDevEntry

				n.NetDevs[SetNetName] = &nd

				if SetNetDev == "" {
					n.NetDevs[SetNetName].Device.Set(SetNetName)
				}
			}
			var def bool = true

			SetNetOnBoot = "yes"

			for _, n := range n.NetDevs {
				if n.Default.GetB() {
					def = false
				}
			}

			if def {
				SetNetDefault = "yes"
			}

		}

		if SetNetDev != "" {
			if SetNetName == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netname' option\n")
				os.Exit(1)
			}

			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting net Device to: %s\n", n.Id.Get(), SetNetName, SetNetDev)
			n.NetDevs[SetNetName].Device.Set(SetNetDev)
		}

		if SetIpaddr != "" {
			if SetNetName == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netname' option\n")
				os.Exit(1)
			}

			NewIpaddr := util.IncrementIPv4(SetIpaddr, count)

			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting Ipaddr to: %s\n", n.Id.Get(), SetNetName, NewIpaddr)
			n.NetDevs[SetNetName].Ipaddr.Set(NewIpaddr)
		}

		if SetNetmask != "" {
			if SetNetName == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netname' option\n")
				os.Exit(1)
			}

			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting netmask to: %s\n", n.Id.Get(), SetNetName, SetNetmask)
			n.NetDevs[SetNetName].Netmask.Set(SetNetmask)
		}

		if SetGateway != "" {
			if SetNetName == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netname' option\n")
				os.Exit(1)
			}

			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting gateway to: %s\n", n.Id.Get(), SetNetName, SetGateway)
			n.NetDevs[SetNetName].Gateway.Set(SetGateway)
		}

		if SetHwaddr != "" {
			if SetNetName == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netname' option\n")
				os.Exit(1)
			}

			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting HW address to: %s\n", n.Id.Get(), SetNetName, SetHwaddr)
			n.NetDevs[SetNetName].Hwaddr.Set(strings.ToLower(SetHwaddr))
		}

		if SetType != "" {
			if SetNetName == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netname' option\n")
				os.Exit(1)
			}

			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting Type %s\n", n.Id.Get(), SetNetName, SetType)
			n.NetDevs[SetNetName].Type.Set(SetType)
		}

		if SetNetOnBoot != "" {
			if SetNetName == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netname' option\n")
				os.Exit(1)
			}

			if SetNetOnBoot == "yes" || SetNetOnBoot == "y" || SetNetOnBoot == "1" || SetNetOnBoot == "true" {
				wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting ONBOOT\n", n.Id.Get(), SetNetName)
				n.NetDevs[SetNetName].OnBoot.SetB(true)
			} else {
				wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Unsetting ONBOOT\n", n.Id.Get(), SetNetName)
				n.NetDevs[SetNetName].OnBoot.SetB(false)
			}
		}

		if SetNetDefault != "" {
			if SetNetName == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netname' option\n")
				os.Exit(1)
			}

			if SetNetDefault == "yes" || SetNetDefault == "y" || SetNetDefault == "1" || SetNetDefault == "true" {

				// Set all other devices to non-default
				for _, n := range n.NetDevs {
					n.Default.SetB(false)
				}

				wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting DEFAULT\n", n.Id.Get(), SetNetName)
				n.NetDevs[SetNetName].Default.SetB(true)
			} else {
				wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Unsetting DEFAULT\n", n.Id.Get(), SetNetName)
				n.NetDevs[SetNetName].Default.SetB(false)
			}
		}

		if SetNetDevDel {
			if SetNetName == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netname' option\n")
				os.Exit(1)
			}

			if _, ok := n.NetDevs[SetNetName]; !ok {
				wwlog.Printf(wwlog.ERROR, "Network device name doesn't exist: %s\n", SetNetName)
				os.Exit(1)
			}

			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Deleting network device: %s\n", n.Id.Get(), SetNetName)
			delete(n.NetDevs, SetNetName)
		}

		if len(SetTags) > 0 {
			for _, t := range SetTags {
				keyval := strings.SplitN(t, "=", 2)
				key := keyval[0]
				val := keyval[1]

				if _, ok := n.Tags[key]; !ok {
					var nd node.Entry
					n.Tags[key] = &nd
				}

				wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting Tag '%s'='%s'\n", n.Id.Get(), key, val)
				n.Tags[key].Set(val)
			}
		}
		if len(SetDelTags) > 0 {
			for _, t := range SetDelTags {
				keyval := strings.SplitN(t, "=", 1)
				key := keyval[0]

				if _, ok := n.Tags[key]; !ok {
					wwlog.Printf(wwlog.WARN, "Key does not exist: %s\n", key)
					os.Exit(1)
				}

				wwlog.Printf(wwlog.VERBOSE, "Node: %s, Deleting tag: %s\n", n.Id.Get(), key)
				delete(n.Tags, key)
			}
		}

		err := nodeDB.NodeUpdate(n)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}

		count++
	}

	if SetYes {
		err := nodeDB.Persist()
		if err != nil {
			return errors.Wrap(err, "failed to persist nodedb")
		}

		err = warewulfd.DaemonReload()
		if err != nil {
			return errors.Wrap(err, "failed to reload warewulf daemon")
		}
	} else {
		q := fmt.Sprintf("Are you sure you want to modify %d nodes(s)", len(nodes))

		prompt := promptui.Prompt{
			Label:     q,
			IsConfirm: true,
		}

		result, _ := prompt.Run()

		if result == "y" || result == "yes" {
			err := nodeDB.Persist()
			if err != nil {
				return errors.Wrap(err, "failed to persist nodedb")
			}

			err = warewulfd.DaemonReload()
			if err != nil {
				return errors.Wrap(err, "failed to reload warewulf daemon")
			}
		}
	}

	return nil
}
