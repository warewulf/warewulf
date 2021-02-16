package delete

import (
	"fmt"
	"os"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var count int
	var numNodes int
	var numGroups int

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Failed to open node database: %s\n", err)
		os.Exit(1)
	}

	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not load all nodes: %s\n", err)
		os.Exit(1)
	}

	for _, p := range args {
		err := nodeDB.DelProfile(p)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			continue
		}
		for _, n := range nodes {
			for _, np := range n.Profiles {
				if np == p {
					numNodes++
					wwlog.Printf(wwlog.VERBOSE, "Removing profile from node %s: %s\n", n.Id.Get(), p)
					n.Profiles = util.SliceRemoveElement(n.Profiles, p)
					nodeDB.NodeUpdate(n)
				}
			}
		}
		count++
	}

	if count > 0 {
		q := fmt.Sprintf("Are you sure you want to delete %d profile(s) (%d groups, %d nodes)", count, numGroups, numNodes)

		prompt := promptui.Prompt{
			Label:     q,
			IsConfirm: true,
		}

		result, _ := prompt.Run()

		if result == "y" || result == "yes" {
			nodeDB.Persist()
		}

	} else {
		wwlog.Printf(wwlog.INFO, "No groups found\n")
	}

	return nil
}
