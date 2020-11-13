package build

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/build/kernel"
	runtime_overlay "github.com/hpcng/warewulf/internal/app/wwctl/build/runtime-overlay"
	system_overlay "github.com/hpcng/warewulf/internal/app/wwctl/build/system-overlay"
	"github.com/hpcng/warewulf/internal/app/wwctl/build/vnfs"
	"github.com/hpcng/warewulf/internal/pkg/assets"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
)


func CobraRunE(cmd *cobra.Command, args []string) error {
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