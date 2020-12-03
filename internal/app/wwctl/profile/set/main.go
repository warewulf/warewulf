package set

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"os"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var err error
	var profiles []node.NodeInfo

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	if len(args) == 0 {
		args = append(args, "default")
	}

	if SetAll == true {
		profiles, err = nodeDB.FindAllProfiles()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}
	} else {
		var tmp []node.NodeInfo
		tmp, err = nodeDB.FindAllProfiles()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}

		for _, a := range args {
			for _, p := range tmp {
				if p.Id.Get() == a {
					profiles = append(profiles, p)
				}
			}
		}
	}

	for _, p := range profiles {

		if SetComment != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting comment to: %s\n", p.Id, SetComment)

			p.Comment.Set(SetComment)
			err := nodeDB.ProfileUpdate(p)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetDomainName != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting domain name to: %s\n", p.Id, SetDomainName)

			p.DomainName.Set(SetDomainName)
			err := nodeDB.ProfileUpdate(p)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetVnfs != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting VNFS to: %s\n", p.Id, SetVnfs)

			p.Vnfs.Set(SetVnfs)
			err := nodeDB.ProfileUpdate(p)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetKernel != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting Kernel version to: %s\n", p.Id, SetKernel)

			p.KernelVersion.Set(SetKernel)
			err := nodeDB.ProfileUpdate(p)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetIpxe != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting iPXE template to: %s\n", p.Id, SetIpxe)

			p.Ipxe.Set(SetIpxe)
			err := nodeDB.ProfileUpdate(p)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetRuntimeOverlay != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting runtime overlay to: %s\n", p.Id, SetRuntimeOverlay)

			p.RuntimeOverlay.Set(SetRuntimeOverlay)
			err := nodeDB.ProfileUpdate(p)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetSystemOverlay != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting system overlay to: %s\n", p.Id, SetSystemOverlay)

			p.SystemOverlay.Set(SetSystemOverlay)
			err := nodeDB.ProfileUpdate(p)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetIpmiNetmask != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting IPMI username to: %s\n", p.Id, SetIpmiNetmask)

			p.IpmiNetmask.Set(SetIpmiNetmask)
			err := nodeDB.ProfileUpdate(p)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetIpmiUsername != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting IPMI username to: %s\n", p.Id, SetIpmiUsername)

			p.IpmiUserName.Set(SetIpmiUsername)
			err := nodeDB.ProfileUpdate(p)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetIpmiPassword != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting IPMI username to: %s\n", p.Id, SetIpmiPassword)

			p.IpmiPassword.Set(SetIpmiPassword)
			err := nodeDB.ProfileUpdate(p)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}

	}

	if len(profiles) > 0 {
		q := fmt.Sprintf("Are you sure you want to modify %d profile(s)", len(profiles))

		prompt := promptui.Prompt{
			Label:     q,
			IsConfirm: true,
		}

		result, _ := prompt.Run()

		if result == "y" || result == "yes" {
			nodeDB.Persist()
		}

	} else {
		fmt.Printf("No profiles found\n")
	}

	return nil
}
