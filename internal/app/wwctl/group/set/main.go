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
//			err := nodeDB.SetGroupVal(g.Id.String(), "domain", SetDomainName)
//			if err != nil {
//				wwlog.Printf(wwlog.ERROR, "%s\n", err)
//				os.Exit(1)
//			}

//			if SetClearNodes == true {
//				nodes, err := nodeDB.FindAllNodes()
//				if err != nil {
//					wwlog.Printf(wwlog.ERROR, "%s\n", err)
//					os.Exit(1)
//				}
//				for _, n := range nodes {
//					_ = nodeDB.SetNodeVal(g.Id.String(), n.Id.String(), "domain", "")
//				}
//			}
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
