package build

import (
	"os"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var updateNodes []node.NodeInfo

	n, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	if len(args) > 0 && !BuildAll {
		nodes, err := n.FindAllNodes()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not get node list: %s\n", err)
			os.Exit(1)
		}

		for _, node := range nodes {
			if SystemOverlay && node.SystemOverlay.Get() == args[0] {
				updateNodes = append(updateNodes, node)
			} else if node.RuntimeOverlay.Get() == args[0] {
				updateNodes = append(updateNodes, node)
			}
		}
	} else {
		var err error
		updateNodes, err = n.FindAllNodes()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not get node list: %s\n", err)
			os.Exit(1)
		}
	}

	wwlog.Printf(wwlog.DEBUG, "Checking on system overlay update\n")
	if SystemOverlay || BuildAll {
		wwlog.Printf(wwlog.INFO, "Updating System Overlays...\n")
		err := overlay.BuildSystemOverlay(updateNodes)
		if err != nil {
			wwlog.Printf(wwlog.WARN, "Some system overlays failed to be generated: %s\n", err)
		}
	}

	wwlog.Printf(wwlog.DEBUG, "Checking on system overlay update\n")
	if !SystemOverlay || BuildAll {
		wwlog.Printf(wwlog.INFO, "Updating Runtime Overlays...\n")
		err := overlay.BuildRuntimeOverlay(updateNodes)
		if err != nil {
			wwlog.Printf(wwlog.WARN, "Some runtime overlays failed to be generated\n")
		}
	}

	return nil
}
