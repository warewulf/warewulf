package build

import (
	"github.com/hpcng/warewulf/internal/pkg/kernel"
	"github.com/hpcng/warewulf/internal/pkg/node"
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
			if node.KernelVersion.Defined() == true {
				set[node.KernelVersion.Get()]++
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
			wwlog.Printf(wwlog.DEBUG, "evaluating node/kernel: %s/%s\n", node.Fqdn.Get(), node.KernelVersion.Get())
			if node.KernelVersion.Defined() == true {
				set[node.KernelVersion.Get()]++
			}
		}

	} else if len(args) == 1 {
		set[args[0]]++
	} else {
		cmd.Usage()
		os.Exit(1)
	}

	for k := range set {
		wwlog.Printf(wwlog.INFO, "Building kernel: %s\n", k)
		err := kernel.Build(k)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}
	}

	return nil
}
