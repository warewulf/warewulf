package delete

import (
	"fmt"
	"os"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/hpcng/warewulf/pkg/hostlist"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var count int
	var nodeList []node.NodeInfo

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Failed to open node database: %s\n", err)
		os.Exit(1)
	}

	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not get node list: %s\n", err)
		os.Exit(1)
	}

	args = hostlist.Expand(args)

	for _, r := range args {
		var match bool
		for _, n := range nodes {
			if n.Id.Get() == r {
				nodeList = append(nodeList, n)
				match = true
			}
		}

		if !match {
			fmt.Fprintf(os.Stderr, "ERROR: No match for node: %s\n", r)
		}
	}

	if len(nodeList) == 0 {
		fmt.Printf("No nodes found\n")
		os.Exit(1)
	}

	for _, n := range nodeList {
		err := nodeDB.DelNode(n.Id.Get())
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
		} else {
			count++
			fmt.Printf("Deleting node: %s\n", n.Id.Print())
		}
	}

	if SetYes {
		err := nodeDB.Persist()
		if err != nil {
			return errors.Wrap(err, "failed to persist nodedb")
		}
	} else {
		q := fmt.Sprintf("Are you sure you want to delete %d nodes(s)", count)

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
		}
	}

	return nil
}
