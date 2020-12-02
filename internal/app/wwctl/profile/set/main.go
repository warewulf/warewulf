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
	var profiles []node.ProfileInfo

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Cloud not get nodeList: %s\n", err)
		os.Exit(1)
	}

	if len(args) == 0 {
		args = append(args, "default")
	}

	if SetAll == true {
		var tmp []node.ProfileInfo
		tmp, err = nodeDB.FindAllProfiles()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}

		for _, p := range tmp {
			profiles = append(profiles, p)
		}

	} else {
		var tmp []node.ProfileInfo
		tmp, err = nodeDB.FindAllProfiles()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}

		for _, a := range args {
			for _, p := range tmp {
				if p.Id == a {
					profiles = append(profiles, p)
				}
			}
		}
	}

	for _, p := range profiles {

		if SetDomainName != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting domain name to: %s\n", p.Id, SetDomainName)

			p.DomainName = SetDomainName
			err := nodeDB.ProfileUpdate(p)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetVnfs != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting VNFS to: %s\n", p.Id, SetVnfs)

			p.Vnfs = SetVnfs
			err := nodeDB.ProfileUpdate(p)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetKernel != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting Kernel version to: %s\n", p.Id, SetKernel)

			p.KernelVersion = SetKernel
			err := nodeDB.ProfileUpdate(p)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetIpxe != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting iPXE template to: %s\n", p.Id, SetIpxe)

			p.Ipxe = SetIpxe
			err := nodeDB.ProfileUpdate(p)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetRuntimeOverlay != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting runtime overlay to: %s\n", p.Id, SetRuntimeOverlay)

			p.RuntimeOverlay = SetRuntimeOverlay
			err := nodeDB.ProfileUpdate(p)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetSystemOverlay != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting system overlay to: %s\n", p.Id, SetSystemOverlay)

			p.SystemOverlay = SetSystemOverlay
			err := nodeDB.ProfileUpdate(p)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetIpmiNetmask != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting IPMI username to: %s\n", p.Id, SetIpmiNetmask)

			p.IpmiNetmask = SetIpmiNetmask
			err := nodeDB.ProfileUpdate(p)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetIpmiUsername != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting IPMI username to: %s\n", p.Id, SetIpmiUsername)

			p.IpmiUserName = SetIpmiUsername
			err := nodeDB.ProfileUpdate(p)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetIpmiPassword != "" {
			wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Setting IPMI username to: %s\n", p.Id, SetIpmiPassword)

			p.IpmiPassword = SetIpmiPassword
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
