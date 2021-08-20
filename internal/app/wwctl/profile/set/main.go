package set

import (
	"fmt"
	"os"

	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var err error

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	profiles, err := nodeDB.FindAllProfiles()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	if !SetAll {
		if len(args) > 0 {
			profiles = node.FilterByName(profiles, args)
		} else {
			cmd.Usage()
			os.Exit(1)
		}
	}

	if len(profiles) == 0 {
		fmt.Printf("No profiles found\n")
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
			wwlog.Printf(wwlog.ERROR, "Container name does not exist: %s\n", SetContainer)
			if SetForce == false {
				os.Exit(1)
			}
		}
	}

	for _, p := range profiles {
		wwlog.Printf(wwlog.VERBOSE, "Modifying profile: %s\n", p.Id.Get())

		if SetComment != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting comment to: %s\n", p.Id, SetComment)
			p.Comment.Set(SetComment)
		}

		if SetClusterName != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting cluster name to: %s\n", p.Id, SetClusterName)
			p.ClusterName.Set(SetClusterName)
		}

		if SetContainer != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting Container name to: %s\n", p.Id, SetContainer)
			p.ContainerName.Set(SetContainer)
		}

		if SetInit != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting init command to: %s\n", p.Id, SetInit)
			p.Init.Set(SetInit)
		}

		if SetRoot != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting root to: %s\n", p.Id, SetRoot)
			p.Root.Set(SetRoot)
		}

		if SetKernel != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting Kernel to: %s\n", p.Id, SetKernel)
			p.KernelVersion.Set(SetKernel)
		}

		if SetKernelArgs != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting Kernel args to: %s\n", p.Id, SetKernelArgs)
			p.KernelArgs.Set(SetKernelArgs)
		}

		if SetIpxe != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting iPXE template to: %s\n", p.Id, SetIpxe)
			p.Ipxe.Set(SetIpxe)
		}

		if SetRuntimeOverlay != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting runtime overlay to: %s\n", p.Id, SetRuntimeOverlay)
			p.RuntimeOverlay.Set(SetRuntimeOverlay)
		}

		if SetSystemOverlay != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting system overlay to: %s\n", p.Id, SetSystemOverlay)
			p.SystemOverlay.Set(SetSystemOverlay)
		}

		if SetIpmiNetmask != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting IPMI username to: %s\n", p.Id, SetIpmiNetmask)
			p.IpmiNetmask.Set(SetIpmiNetmask)
		}

		if SetIpmiGateway != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting IPMI username to: %s\n", p.Id, SetIpmiGateway)
			p.IpmiGateway.Set(SetIpmiGateway)
		}

		if SetIpmiUsername != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting IPMI username to: %s\n", p.Id, SetIpmiUsername)
			p.IpmiUserName.Set(SetIpmiUsername)
		}

		if SetIpmiPassword != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting IPMI username to: %s\n", p.Id, SetIpmiPassword)
			p.IpmiPassword.Set(SetIpmiPassword)
		}

		if SetIpmiInterface != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting IPMI username to: %s\n", p.Id, SetIpmiInterface)
			p.IpmiInterface.Set(SetIpmiInterface)
		}

		if SetNetDevDel == true {
			if SetNetDev == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netdev' option\n")
				os.Exit(1)
			}

			if _, ok := p.NetDevs[SetNetDev]; !ok {
				wwlog.Printf(wwlog.ERROR, "Profile '%s': network Device doesn't exist: %s\n", p.Id.Get(), SetNetDev)
				os.Exit(1)
			}

			wwlog.Printf(wwlog.VERBOSE, "Profile %s: Deleting network device: %s\n", p.Id.Get(), SetNetDev)
			delete(p.NetDevs, SetNetDev)
		}

		if SetIpaddr != "" {
			if SetNetDev == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netdev' option\n")
				os.Exit(1)
			}

			if _, ok := p.NetDevs[SetNetDev]; !ok {
				var nd node.NetDevEntry
				p.NetDevs[SetNetDev] = &nd
			}

			wwlog.Printf(wwlog.VERBOSE, "Profile '%s': Setting IP address to: %s:%s\n", p.Id.Get(), SetNetDev, SetHwaddr)
			p.NetDevs[SetNetDev].Ipaddr.Set(SetIpaddr)
		}

		if SetNetmask != "" {
			if SetNetDev == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netdev' option\n")
				os.Exit(1)
			}

			if _, ok := p.NetDevs[SetNetDev]; !ok {
				var nd node.NetDevEntry
				p.NetDevs[SetNetDev] = &nd
			}

			wwlog.Printf(wwlog.VERBOSE, "Profile '%s': Setting netmask to: %s:%s\n", p.Id.Get(), SetNetDev, SetHwaddr)
			p.NetDevs[SetNetDev].Netmask.Set(SetNetmask)
		}

		if SetGateway != "" {
			if SetNetDev == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netdev' option\n")
				os.Exit(1)
			}

			if _, ok := p.NetDevs[SetNetDev]; !ok {
				var nd node.NetDevEntry
				p.NetDevs[SetNetDev] = &nd
			}

			wwlog.Printf(wwlog.VERBOSE, "Profile '%s': Setting gateway to: %s:%s\n", p.Id.Get(), SetNetDev, SetHwaddr)
			p.NetDevs[SetNetDev].Gateway.Set(SetGateway)
		}

		if SetHwaddr != "" {
			if SetNetDev == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netdev' option\n")
				os.Exit(1)
			}

			if _, ok := p.NetDevs[SetNetDev]; !ok {
				var nd node.NetDevEntry
				p.NetDevs[SetNetDev] = &nd
			}

			wwlog.Printf(wwlog.VERBOSE, "Profile '%s': Setting HW address to: %s:%s\n", p.Id.Get(), SetNetDev, SetHwaddr)
			p.NetDevs[SetNetDev].Hwaddr.Set(SetHwaddr)
		}

		if SetType != "" {
			if SetNetDev == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netdev' option\n")
				os.Exit(1)
			}

			if _, ok := p.NetDevs[SetNetDev]; !ok {
				var nd node.NetDevEntry
				p.NetDevs[SetNetDev] = &nd
			}

			wwlog.Printf(wwlog.VERBOSE, "Profile '%s': Setting HW address to: %s:%s\n", p.Id.Get(), SetNetDev, SetType)
			p.NetDevs[SetNetDev].Type.Set(SetType)
		}

		if SetNetDevDefault == true {
			if SetNetDev == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netdev' option\n")
				os.Exit(1)
			}

			if _, ok := p.NetDevs[SetNetDev]; !ok {
				var nd node.NetDevEntry
				p.NetDevs[SetNetDev] = &nd
			}

			wwlog.Printf(wwlog.VERBOSE, "Profile: %s:%s, Setting device as default\n", p.Id.Get(), SetNetDev)
			for _, dev := range p.NetDevs {
				// First clear all other devices that might be configured as default
				dev.Default.SetB(false)
			}
			p.NetDevs[SetNetDev].Default.SetB(true)
		}

		if SetValue != "" {
			if SetKey == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--key/-k' option\n")
				os.Exit(1)
			}

			if _, ok := p.Keys[SetKey]; !ok {
				var nd node.Entry
				p.Keys[SetKey] = &nd
			}
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s:%s, Setting Value %s\n", p.Id.Get(), SetKey, SetValue)
			p.Keys[SetKey].Set(SetValue)
		}

		if SetKeyDel == true {
			if SetKey == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--key/-k' option\n")
				os.Exit(1)
			}

			if _, ok := p.Keys[SetKey]; !ok {
				wwlog.Printf(wwlog.ERROR, "Custom key doesn't exist: %s\n", SetKey)
				os.Exit(1)
			}

			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Deleting custom key: %s\n", p.Id.Get(), SetNetDev)
			delete(p.Keys, SetKey)
		}

		err := nodeDB.ProfileUpdate(p)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}
	}

	if len(profiles) > 0 {
		if SetYes == true {
			nodeDB.Persist()
			warewulfd.DaemonReload()
		} else {
			q := fmt.Sprintf("Are you sure you want to modify %d profile(s)", len(profiles))

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
		fmt.Printf("No profiles found\n")
	}

	return nil
}
