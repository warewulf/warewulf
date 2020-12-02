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
	var groups []node.GroupInfo

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Cloud not get nodeList: %s\n", err)
		os.Exit(1)
	}

	if SetGroupAll == true {
		var tmp []node.GroupInfo
		tmp, err = nodeDB.FindAllGroups()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}

		for _, g := range tmp {
			groups = append(groups, g)
		}

	} else if len(args) > 0 {
		var tmp []node.GroupInfo
		tmp, err = nodeDB.FindAllGroups()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}

		for _, a := range args {
			for _, g := range tmp {
				if g.Id == a {
					groups = append(groups, g)
				}
			}
		}

	} else {
		cmd.Usage()
		os.Exit(1)
	}

	for _, g := range groups {

		if SetDomainName != "" {
			wwlog.Printf(wwlog.VERBOSE, "Group: %s, Setting domain name to: %s\n", g.Id, SetDomainName)

			g.DomainName = SetDomainName
			err := nodeDB.GroupUpdate(g)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetVnfs != "" {
			wwlog.Printf(wwlog.VERBOSE, "Group: %s, Setting VNFS to: %s\n", g.Id, SetVnfs)

			g.Vnfs = SetVnfs
			err := nodeDB.GroupUpdate(g)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetKernel != "" {
			wwlog.Printf(wwlog.VERBOSE, "Group: %s, Setting kernel to: %s\n", g.Id, SetKernel)

			g.KernelVersion = SetKernel
			err := nodeDB.GroupUpdate(g)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetIpmiNetmask != "" {
			wwlog.Printf(wwlog.VERBOSE, "Group: %s, Setting IPMI username to: %s\n", g.Id, SetIpmiNetmask)

			g.IpmiNetmask = SetIpmiNetmask
			err := nodeDB.GroupUpdate(g)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetIpmiUsername != "" {
			wwlog.Printf(wwlog.VERBOSE, "Group: %s, Setting IPMI username to: %s\n", g.Id, SetIpmiUsername)

			g.IpmiUserName = SetIpmiUsername
			err := nodeDB.GroupUpdate(g)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetIpmiPassword != "" {
			wwlog.Printf(wwlog.VERBOSE, "Group: %s, Setting IPMI password to: %s\n", g.Id, SetIpmiPassword)

			g.IpmiPassword = SetIpmiPassword
			err := nodeDB.GroupUpdate(g)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetSystemOverlay != "" {
			wwlog.Printf(wwlog.VERBOSE, "Group: %s, Setting system overlay to: %s\n", g.Id, SetSystemOverlay)

			g.SystemOverlay = SetSystemOverlay
			err := nodeDB.GroupUpdate(g)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetRuntimeOverlay != "" {
			wwlog.Printf(wwlog.VERBOSE, "Group: %s, Setting runtime overlay to: %s\n", g.Id, SetRuntimeOverlay)

			g.RuntimeOverlay = SetRuntimeOverlay
			err := nodeDB.GroupUpdate(g)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}

		if len(SetAddProfile) > 0 {
			for _, p := range SetAddProfile {
				wwlog.Printf(wwlog.VERBOSE, "Adding profile to '%s': '%s'\n", g.Id, p)
				g.Profiles = util.SliceAddUniqueElement(g.Profiles, p)
			}
			err := nodeDB.GroupUpdate(g)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if len(SetDelProfile) > 0 {
			for _, p := range SetDelProfile {
				wwlog.Printf(wwlog.VERBOSE, "Removing profile to '%s': '%s'\n", g.Id, p)
				g.Profiles = util.SliceRemoveElement(g.Profiles, p)
			}
			err := nodeDB.GroupUpdate(g)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}

	}

	if len(groups) > 0 {
		q := fmt.Sprintf("Are you sure you want to modify %d group(s)", len(groups))

		prompt := promptui.Prompt{
			Label:     q,
			IsConfirm: true,
		}

		result, _ := prompt.Run()

		if result == "y" || result == "yes" {
			nodeDB.Persist()
		}

	} else {
		fmt.Printf("No groups found\n")
	}

	return nil
}
