package delete

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"os"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var count int
	var numNodes int

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Failed to open node database: %s\n", err)
		os.Exit(1)
	}

	nodes, err := nodeDB. FindAllNodes()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not load all nodes: %s\n", err)
		os.Exit(1)
	}

	for _, g := range args {
		err := nodeDB.DelGroup(g)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
		} else {
			for _, n := range nodes {
				if n.GroupName == g {
					numNodes ++
				}
			}
			count ++
		}
	}

	if count > 0 {
		q := fmt.Sprintf("Are you sure you want to delete %d group(s) (%d nodes)", count, numNodes)

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
