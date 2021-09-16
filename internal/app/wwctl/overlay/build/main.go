package build

import (
	"os"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var updateNodes []node.NodeInfo

	overlayKind := args[0]
	overlayName := args[1]

	if overlayKind != "system" && overlayKind != "runtime" {
		return errors.New("overlay kind must be of type 'system' or 'runtime'")
	}

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
			if overlayKind == "system" && node.SystemOverlay.Get() == overlayName {
				updateNodes = append(updateNodes, node)
			} else if overlayKind == "runtime" && node.RuntimeOverlay.Get() == overlayName {
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
	if overlayKind == "system" || BuildAll {
		wwlog.Printf(wwlog.INFO, "Updating System Overlays...\n")
		err := overlay.BuildSystemOverlay(updateNodes)
		if err != nil {
			wwlog.Printf(wwlog.WARN, "Some system overlays failed to be generated: %s\n", err)
		}
	}

	wwlog.Printf(wwlog.DEBUG, "Checking on runtime overlay update\n")
	if overlayKind == "runtime" || BuildAll {
		wwlog.Printf(wwlog.INFO, "Updating Runtime Overlays...\n")
		err := overlay.BuildRuntimeOverlay(updateNodes)
		if err != nil {
			wwlog.Printf(wwlog.WARN, "Some runtime overlays failed to be generated\n")
		}
	}

	return nil
}
