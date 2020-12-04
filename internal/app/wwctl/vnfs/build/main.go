package build

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/vnfs"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var nodes []node.NodeInfo
	set := make(map[string]int)

	n, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	if len(args) == 1 && ByNode == true {
		var err error
		nodes, err = n.SearchByName(args[0])
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not find nodes for search term: %s\n", args[0])
			os.Exit(1)
		}

		for _, node := range nodes {
			if node.Vnfs.Defined() == true {
				set[node.Vnfs.Get()]++
			}
		}

	} else if BuildAll == true {
		var err error
		nodes, err = n.FindAllNodes()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not get list of nodes: %s\n", err)
			os.Exit(1)
		}

		for _, node := range nodes {
			if node.Vnfs.Defined() == true {
				wwlog.Printf(wwlog.VERBOSE, "Adding VNFS to list: %s (%s)\n", node.Vnfs.Get(), node.Id.Get())
				set[node.Vnfs.Get()]++
			}
		}

	} else if len(args) == 1 {
		set[args[0]]++
	} else {
		cmd.Usage()
		os.Exit(1)
	}

	for v := range set {
		fmt.Printf("Building VNFS: %s\n", v)
		err := vnfs.Build(v, BuildForce)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}
	}

	return nil
}
