package build

import (
	"github.com/hpcng/warewulf/internal/pkg/assets"
	"github.com/hpcng/warewulf/internal/pkg/kernel"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/vnfs"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
)


func CobraRunE(cmd *cobra.Command, args []string) error {
	var nodes []assets.NodeInfo
	showHelp := true

	if len(args) > 0 {
		var err error
		nodes, err = assets.SearchByName(args[0])
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not find nodes for search term: %s\n", args[0])
			os.Exit(1)
		}

	} else {
		var err error
		nodes, err = assets.FindAllNodes()
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
			set[node.Vnfs] ++
		}
		for e := range set {
			vnfs.Build(e, buildForce)
		}
	}

	if buildKernel == true || buildAll == true {
		set := make(map[string]int)
		showHelp = false

		wwlog.Printf(wwlog.INFO, "Building Kernel images...\n")

		for _, node := range nodes {
			set[node.KernelVersion] ++
		}
		for e := range set {
			kernel.Build(e)
		}
	}

	if buildSystemOverlay == true || buildAll == true {
		wwlog.Printf(wwlog.INFO, "Building System Overlays...\n")
		showHelp = false

		overlay.SystemBuild(nodes, buildForce)
	}

	if buildRuntimeOverlay == true || buildAll == true {
		wwlog.Printf(wwlog.INFO, "Building Runtime Overlays...n")
		showHelp = false

		overlay.RuntimeBuild(nodes, buildForce)
	}

	if showHelp == true {
		cmd.Usage()
		os.Exit(1)
	}

	return nil
}


/*
func CobraRunEA(cmd *cobra.Command, args []string) error {
	var nodeList []assets.NodeInfo

	if buildAll == true {
		wwlog.Printf(wwlog.VERBOSE, "Building all components\n")
		buildVnfs = true
		buildKernel = true
		buildSystemOverlay = true;
		buildRuntimeOverlay = true;
	}

	if len(args) >= 1 {
		nodeList, _ = assets.SearchByName(args[0])
	} else {
		nodeList, _ = assets.FindAllNodes()
	}

	if len(nodeList) == 0 {
		wwlog.Printf(wwlog.ERROR, "No nodes found matching: '%s'\n", args[0])
		os.Exit(255)
	} else {
		wwlog.Printf(wwlog.VERBOSE, "Found matching nodes for build: %d\n", len(nodeList))
	}

	if buildVnfs == true {
//		wwlog.Printf(wwlog.INFO, "===============================================================================\n")

		err := vnfs.Build(nodeList, buildForce)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(255)
		}
	}

	if buildKernel == true {
//		wwlog.Printf(wwlog.INFO, "===============================================================================\n")

		err := kernel.Build(nodeList, buildForce)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(255)
		}
	}

	if buildSystemOverlay == true {
//		wwlog.Printf(wwlog.INFO, "===============================================================================\n")
		err := system_overlay.Build(nodeList, buildForce)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(255)
		}
	}

	if buildRuntimeOverlay == true {
//		wwlog.Printf(wwlog.INFO, "===============================================================================\n")
		err := runtime_overlay.Build(nodeList, buildForce)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(255)
		}
	}


	return nil
}

 */