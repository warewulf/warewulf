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

	if len(args) == 0 {
		args = append(args, "default")
	}

	if SetGroupAll == true {
		groups, err = nodeDB.FindAllGroups()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}

	} else {
		var tmp []node.GroupInfo
		tmp, err = nodeDB.FindAllGroups()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}

		for _, a := range args {
			for _, g := range tmp {
				if g.Id.Get() == a {
					groups = append(groups, g)
				}
			}
		}
	}

	for _, g := range groups {

		if SetDomainName != "" {
			wwlog.Printf(wwlog.VERBOSE, "Group: %s, Setting domain name to: %s\n", g.Id, SetDomainName)

			g.DomainName.SetGroup(SetDomainName)
			err := nodeDB.GroupUpdate(g)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetVnfs != "" {
			wwlog.Printf(wwlog.VERBOSE, "Group: %s, Setting VNFS to: %s\n", g.Id, SetVnfs)

			g.Vnfs.SetGroup(SetVnfs)
			err := nodeDB.GroupUpdate(g)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetKernel != "" {
			wwlog.Printf(wwlog.VERBOSE, "Group: %s, Setting kernel to: %s\n", g.Id, SetKernel)

			g.KernelVersion.SetGroup(SetKernel)
			err := nodeDB.GroupUpdate(g)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetIpmiNetmask != "" {
			wwlog.Printf(wwlog.VERBOSE, "Group: %s, Setting IPMI username to: %s\n", g.Id, SetIpmiNetmask)

			g.IpmiNetmask.SetGroup(SetIpmiNetmask)
			err := nodeDB.GroupUpdate(g)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetIpmiUsername != "" {
			wwlog.Printf(wwlog.VERBOSE, "Group: %s, Setting IPMI username to: %s\n", g.Id, SetIpmiUsername)

			g.IpmiUserName.SetGroup(SetIpmiUsername)
			err := nodeDB.GroupUpdate(g)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetIpmiPassword != "" {
			wwlog.Printf(wwlog.VERBOSE, "Group: %s, Setting IPMI password to: %s\n", g.Id, SetIpmiPassword)

			g.IpmiPassword.SetGroup(SetIpmiPassword)
			err := nodeDB.GroupUpdate(g)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetSystemOverlay != "" {
			wwlog.Printf(wwlog.VERBOSE, "Group: %s, Setting system overlay to: %s\n", g.Id, SetSystemOverlay)

			g.SystemOverlay.SetGroup(SetSystemOverlay)
			err := nodeDB.GroupUpdate(g)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetRuntimeOverlay != "" {
			wwlog.Printf(wwlog.VERBOSE, "Group: %s, Setting runtime overlay to: %s\n", g.Id, SetRuntimeOverlay)

			g.RuntimeOverlay.SetGroup(SetRuntimeOverlay)
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
