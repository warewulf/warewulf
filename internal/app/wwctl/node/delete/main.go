package delete

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	nodeDB, err := node.New()
	if err != nil {
		return fmt.Errorf("failed to open node database: %w", err)
	}

	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		return fmt.Errorf("could not get node list: %w", err)
	}

	nodeArgs := hostlist.Expand(args)
	var nodeList []node.Node
	for _, r := range nodeArgs {
		var match bool
		for _, n := range nodes {
			if n.Id() == r {
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
		return
	}

	if !SetYes {
		yes := util.Confirm(fmt.Sprintf("Are you sure you want to delete %d nodes(s)", len(nodeList)))
		if !yes {
			return
		}
	}

	for _, n := range nodeList {
		if err := nodeDB.DelNode(n.Id()); err != nil {
			wwlog.Error("%s", err)
		} else {
			wwlog.Verbose("Deleting node: %s\n", n.Id())
		}
	}

	if err := nodeDB.Persist(); err != nil {
		return fmt.Errorf("failed to persist nodedb: %w", err)
	}
	return warewulfd.DaemonReload()
}
