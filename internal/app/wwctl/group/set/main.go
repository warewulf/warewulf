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

	if len(args) > 0 {
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
		if SetVnfs != "" {
			wwlog.Printf(wwlog.VERBOSE, "Group: %s, Setting vnfs to: %s\n", g.Id, SetVnfs)
			err := nodeDB.SetGroupVal(g.Id, "vnfs", SetVnfs)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}

			if SetClearNodes == true {
				nodes, err := nodeDB.FindAllNodes()
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "%s\n", err)
					os.Exit(1)
				}
				for _, n := range nodes {
					_ = nodeDB.SetNodeVal(g.Id, n.Id, "vnfs", "")
				}
			}
		}

		if SetKernel != "" {
			wwlog.Printf(wwlog.VERBOSE, "Group: %s, Setting kernel to: %s\n", g.Id, SetVnfs)
			err := nodeDB.SetGroupVal(g.Id, "kernel", SetKernel)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}

			if SetClearNodes == true {
				nodes, err := nodeDB.FindAllNodes()
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "%s\n", err)
					os.Exit(1)
				}
				for _, n := range nodes {
					_ = nodeDB.SetNodeVal(g.Id, n.Id, "kernel", "")
				}
			}
		}

		if SetDomainName != "" {
			wwlog.Printf(wwlog.VERBOSE, "Group: %s, Setting domain name to: %s\n", g.Id, SetDomainName)
			err := nodeDB.SetGroupVal(g.Id, "domain", SetDomainName)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}

			if SetClearNodes == true {
				nodes, err := nodeDB.FindAllNodes()
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "%s\n", err)
					os.Exit(1)
				}
				for _, n := range nodes {
					_ = nodeDB.SetNodeVal(g.Id, n.Id, "domain", "")
				}
			}
		}

		if SetIpxe != "" {
			wwlog.Printf(wwlog.VERBOSE, "Group: %s, Setting iPXE template to: %s\n", g.Id, SetIpxe)
			err := nodeDB.SetGroupVal(g.Id, "ipxe", SetIpxe)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}

			if SetClearNodes == true {
				nodes, err := nodeDB.FindAllNodes()
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "%s\n", err)
					os.Exit(1)
				}
				for _, n := range nodes {
					_ = nodeDB.SetNodeVal(g.Id, n.Id, "ipxe", "")
				}
			}
		}

		if SetRuntimeOverlay != "" {
			wwlog.Printf(wwlog.VERBOSE, "Group: %s, Setting runtime overlay to: %s\n", g.Id, SetRuntimeOverlay)
			err := nodeDB.SetGroupVal(g.Id, "runtimeoverlay", SetRuntimeOverlay)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
			if SetClearNodes == true {
				nodes, err := nodeDB.FindAllNodes()
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "%s\n", err)
					os.Exit(1)
				}
				for _, n := range nodes {
					_ = nodeDB.SetNodeVal(g.Id, n.Id, "runtimeoverlay", "")
				}
			}
		}

		if SetSystemOverlay != "" {
			wwlog.Printf(wwlog.VERBOSE, "Group: %s, Setting system overlay to: %s\n", g.Id, SetSystemOverlay)
			err := nodeDB.SetGroupVal(g.Id, "systemoverlay", SetSystemOverlay)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
			if SetClearNodes == true {
				nodes, err := nodeDB.FindAllNodes()
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "%s\n", err)
					os.Exit(1)
				}
				for _, n := range nodes {
					_ = nodeDB.SetNodeVal(g.Id, n.Id, "systemoverlay", "")
				}
			}
		}
		if SetIpmiIpaddr != "" {
			wwlog.Printf(wwlog.VERBOSE, "Group: %s, Setting IPMI IP address to: %s\n", g.Id, SetIpmiIpaddr)
			err := nodeDB.SetGroupVal(g.Id, "ipmiipaddr", SetIpmiIpaddr)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
			if SetClearNodes == true {
				nodes, err := nodeDB.FindAllNodes()
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "%s\n", err)
					os.Exit(1)
				}
				for _, n := range nodes {
					_ = nodeDB.SetNodeVal(g.Id, n.Id, "ipmiipaddr", "")
				}
			}
		}
		if SetIpmiUsername != "" {
			wwlog.Printf(wwlog.VERBOSE, "Group: %s, Setting IPMI IP username to: %s\n", g.Id, SetIpmiUsername)
			err := nodeDB.SetGroupVal(g.Id, "ipmiusername", SetIpmiUsername)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
			if SetClearNodes == true {
				nodes, err := nodeDB.FindAllNodes()
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "%s\n", err)
					os.Exit(1)
				}
				for _, n := range nodes {
					_ = nodeDB.SetNodeVal(g.Id, n.Id, "ipmiusername", "")
				}
			}
		}
		if SetIpmiPassword != "" {
			wwlog.Printf(wwlog.VERBOSE, "Group: %s, Setting IPMI IP password to: %s\n", g.Id, SetIpmiPassword)
			err := nodeDB.SetGroupVal(g.Id, "ipmipassword", SetIpmiPassword)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
			if SetClearNodes == true {
				nodes, err := nodeDB.FindAllNodes()
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "%s\n", err)
					os.Exit(1)
				}
				for _, n := range nodes {
					_ = nodeDB.SetNodeVal(g.Id, n.Id, "ipmipassword", "")
				}
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
