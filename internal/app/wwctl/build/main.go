package build

import (
	"os"

	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/kernel"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var nodes []node.NodeInfo
	showHelp := true

	n, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	if len(args) > 0 {
		var err error
		nodes, err = n.SearchByName(args[0])
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not find nodes for search term: %s\n", args[0])
			os.Exit(1)
		}

	} else {
		var err error
		nodes, err = n.FindAllNodes()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not get list of nodes: %s\n", err)
			os.Exit(1)
		}

	}

	if buildVnfs == true || buildAll == true {
		set := make(map[string]int)
		showHelp = false

		wwlog.Printf(wwlog.INFO, "Building VNFS images...\n")

		for _, node := range nodes {
			set[node.Vnfs.Get()]++
		}
		for e := range set {
			container.Build(e, buildForce)
		}
	}

	if buildKernel == true || buildAll == true {
		set := make(map[string]int)
		showHelp = false

		wwlog.Printf(wwlog.INFO, "Building Kernel images...\n")

		for _, node := range nodes {
			set[node.KernelVersion.Get()]++
		}
		for e := range set {
			kernel.Build(e)
		}
	}

	if buildSystemOverlay == true || buildAll == true {
		wwlog.Printf(wwlog.INFO, "Building System Overlays...\n")
		showHelp = false

		//		overlay.SystemBuild(nodes, buildForce)
	}

	if buildRuntimeOverlay == true || buildAll == true {
		wwlog.Printf(wwlog.INFO, "Building Runtime Overlays...n")
		showHelp = false

		//		overlay.RuntimeBuild(nodes, buildForce)
	}

	if showHelp == true {
		cmd.Usage()
		os.Exit(1)
	}

	return nil
}
