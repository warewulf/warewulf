package build

import (
	"github.com/hpcng/warewulf/internal/pkg/assets"
	"github.com/hpcng/warewulf/internal/pkg/kernel"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var nodes []assets.NodeInfo
	set := make(map[string]int)

	if len(args) == 1 && ByNode == true {
		var err error
		nodes, err = assets.SearchByName(args[0])
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not find nodes for search term: %s\n", args[0])
			os.Exit(1)
		}

		for _, node := range nodes {
			set[node.KernelVersion] ++
		}

	} else if BuildAll == true {
		var err error
		nodes, err = assets.FindAllNodes()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not get list of nodes: %s\n", err)
			os.Exit(1)
		}

		for _, node := range nodes {
			set[node.KernelVersion] ++
		}

	} else if len(args) == 1 {
		set[args[0]] ++
	} else {
		cmd.Usage()
		os.Exit(1)
	}

	for k := range set {
		err := kernel.Build(k)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}
	}

	return nil
}