package set

import (
	"fmt"
	"os"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
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
			//nolint:errcheck
			cmd.Usage()
			os.Exit(1)
		}
	}

	if len(profiles) == 0 {
		fmt.Printf("No profiles found\n")
		os.Exit(1)
	}

	for _, p := range profiles {
		wwlog.Printf(wwlog.VERBOSE, "Modifying profile: %s\n", p.Id.Get())

		if SetComment != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting comment to: %s\n", p.Id.Get(), SetComment)
			p.Comment.Set(SetComment)
		}

		if SetClusterName != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting cluster name to: %s\n", p.Id.Get(), SetClusterName)
			p.ClusterName.Set(SetClusterName)
		}

		if SetContainer != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting Container name to: %s\n", p.Id.Get(), SetContainer)
			p.ContainerName.Set(SetContainer)
		}

		if SetInit != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting init command to: %s\n", p.Id.Get(), SetInit)
			p.Init.Set(SetInit)
		}

		if SetRoot != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting root to: %s\n", p.Id.Get(), SetRoot)
			p.Root.Set(SetRoot)
		}

		if SetKernel != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting Kernel to: %s\n", p.Id.Get(), SetKernel)
			p.KernelVersion.Set(SetKernel)
		}

		if SetKernelArgs != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting Kernel args to: %s\n", p.Id.Get(), SetKernelArgs)
			p.KernelArgs.Set(SetKernelArgs)
		}

		if SetIpxe != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting iPXE template to: %s\n", p.Id.Get(), SetIpxe)
			p.Ipxe.Set(SetIpxe)
		}

		if SetRuntimeOverlay != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting runtime overlay to: %s\n", p.Id.Get(), SetRuntimeOverlay)
			p.RuntimeOverlay.Set(SetRuntimeOverlay)
		}

		if SetSystemOverlay != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting system overlay to: %s\n", p.Id.Get(), SetSystemOverlay)
			p.SystemOverlay.Set(SetSystemOverlay)
		}

		if SetIpmiNetmask != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting IPMI netmask to: %s\n", p.Id.Get(), SetIpmiNetmask)
			p.IpmiNetmask.Set(SetIpmiNetmask)
		}

		if SetIpmiPort != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting IPMI port to: %s\n", p.Id.Get(), SetIpmiPort)
			p.IpmiPort.Set(SetIpmiPort)
		}

		if SetIpmiGateway != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting IPMI gateway to: %s\n", p.Id.Get(), SetIpmiGateway)
			p.IpmiGateway.Set(SetIpmiGateway)
		}

		if SetIpmiUsername != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting IPMI username to: %s\n", p.Id.Get(), SetIpmiUsername)
			p.IpmiUserName.Set(SetIpmiUsername)
		}

		if SetIpmiPassword != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting IPMI password to: %s\n", p.Id.Get(), SetIpmiPassword)
			p.IpmiPassword.Set(SetIpmiPassword)
		}

		if SetIpmiInterface != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting IPMI interface to: %s\n", p.Id.Get(), SetIpmiInterface)
			p.IpmiInterface.Set(SetIpmiInterface)
		}

		if SetDiscoverable {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting all nodes to discoverable\n", p.Id.Get())
			p.Discoverable.SetB(true)
		}

		if SetUndiscoverable {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting all nodes to undiscoverable\n", p.Id.Get())
			p.Discoverable.SetB(false)
		}

		if SetNetName != "" {
			if _, ok := p.NetDevs[SetNetName]; !ok {
				var nd node.NetDevEntry
				p.NetDevs[SetNetName] = &nd

				SetNetOnBoot = "yes"

				if SetNetDev == "" {
					p.NetDevs[SetNetName].Device.Set(SetNetName)
				}
			}
		}

		if SetNetDev != "" {
			if SetNetName == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netname' option\n")
				os.Exit(1)
			}

			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting net Device to: %s\n", p.Id.Get(), SetNetName, SetNetDev)
			p.NetDevs[SetNetName].Device.Set(SetNetDev)
		}

		if SetIpaddr != "" {
			if SetNetName == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netname' option\n")
				os.Exit(1)
			}

			wwlog.Printf(wwlog.VERBOSE, "Profile '%s': Setting IP address to: %s:%s\n", p.Id.Get(), SetNetName, SetHwaddr)
			p.NetDevs[SetNetName].Ipaddr.Set(SetIpaddr)
		}

		if SetNetmask != "" {
			if SetNetName == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netname' option\n")
				os.Exit(1)
			}

			wwlog.Printf(wwlog.VERBOSE, "Profile '%s': Setting netmask to: %s:%s\n", p.Id.Get(), SetNetName, SetHwaddr)
			p.NetDevs[SetNetName].Netmask.Set(SetNetmask)
		}

		if SetGateway != "" {
			if SetNetName == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netname' option\n")
				os.Exit(1)
			}

			wwlog.Printf(wwlog.VERBOSE, "Profile '%s': Setting gateway to: %s:%s\n", p.Id.Get(), SetNetName, SetHwaddr)
			p.NetDevs[SetNetName].Gateway.Set(SetGateway)
		}

		if SetHwaddr != "" {
			if SetNetName == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netname' option\n")
				os.Exit(1)
			}

			wwlog.Printf(wwlog.VERBOSE, "Profile '%s': Setting HW address to: %s:%s\n", p.Id.Get(), SetNetName, SetHwaddr)
			p.NetDevs[SetNetName].Hwaddr.Set(SetHwaddr)
		}

		if SetType != "" {
			if SetNetName == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netname' option\n")
				os.Exit(1)
			}

			wwlog.Printf(wwlog.VERBOSE, "Profile '%s': Setting HW address to: %s:%s\n", p.Id.Get(), SetNetName, SetType)
			p.NetDevs[SetNetName].Type.Set(SetType)
		}

		if SetNetOnBoot != "" {
			if SetNetName == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netname' option\n")
				os.Exit(1)
			}

			if SetNetOnBoot == "yes" || SetNetOnBoot == "y" || SetNetOnBoot == "1" || SetNetOnBoot == "true" {
				wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting ONBOOT\n", p.Id.Get(), SetNetName)
				p.NetDevs[SetNetName].OnBoot.SetB(true)
			} else {
				wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Unsetting ONBOOT\n", p.Id.Get(), SetNetName)
				p.NetDevs[SetNetName].OnBoot.SetB(false)
			}
		}

		if SetNetDevDel {
			if SetNetDev == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netname' option\n")
				os.Exit(1)
			}

			if _, ok := p.NetDevs[SetNetDev]; !ok {
				wwlog.Printf(wwlog.ERROR, "Profile '%s': network Device doesn't exist: %s\n", p.Id.Get(), SetNetDev)
				os.Exit(1)
			}

			wwlog.Printf(wwlog.VERBOSE, "Profile %s: Deleting network device: %s\n", p.Id.Get(), SetNetDev)
			delete(p.NetDevs, SetNetDev)
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

		if SetKeyDel {
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
			q := fmt.Sprintf("Are you sure you want to modify %d profile(s)", len(profiles))

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
					return errors.Wrap(err, "failed to reload daemon")
				}
			}
		}
	} else {
		fmt.Printf("No profiles found\n")
	}

	return nil
}
