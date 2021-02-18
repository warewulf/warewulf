package delete

import (
	"fmt"
	"os"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var count int
	var numNodes int

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

	for _, c := range args {
		err := nodeDB.DelController(c)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
		} else {
			for _, n := range nodes {
				if n.Cid.Get() == c {
					numNodes++
				}
			}
			count++
		}
	}

	if count > 0 {
		q := fmt.Sprintf("Are you sure you want to delete %d controllers(s) (%d nodes)", count, numNodes)

		prompt := promptui.Prompt{
			Label:     q,
			IsConfirm: true,
		}

		result, _ := prompt.Run()

		if result == "y" || result == "yes" {
			nodeDB.Persist()
		}

	} else {
		wwlog.Printf(wwlog.INFO, "No controllers found\n")
	}

	return nil
}
