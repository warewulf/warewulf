package create

import (
	"os"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	overlayKind := args[0]
	overlayName := args[1]

	if overlayKind != "system" && overlayKind != "runtime" {
		return errors.New("overlay kind must be of type 'system' or 'runtime'")
	}

	if overlayKind == "system" {
		err := overlay.SystemOverlayInit(overlayName)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}
		wwlog.Printf(wwlog.INFO, "Created new system overlay: %s\n", overlayName)
	} else {
		err := overlay.RuntimeOverlayInit(overlayName)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}
		wwlog.Printf(wwlog.INFO, "Created new runtime overlay: %s\n", overlayName)
	}

	if !NoOverlayUpdate {
		n, err := node.New()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
			os.Exit(1)
		}

		nodes, err := n.FindAllNodes()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not get nodeList: %s\n", err)
			os.Exit(1)
		}

		var updateNodes []node.NodeInfo

		for _, node := range nodes {
			if overlayKind == "system" && node.SystemOverlay.Get() == overlayName {
				updateNodes = append(updateNodes, node)
			} else if overlayKind == "runtime" && node.RuntimeOverlay.Get() == overlayName {
				updateNodes = append(updateNodes, node)
			}
		}

		if overlayKind == "system" {
			wwlog.Printf(wwlog.INFO, "Updating System Overlays...\n")
			return overlay.BuildSystemOverlay(updateNodes)
		} else if overlayKind == "runtime" {
			wwlog.Printf(wwlog.INFO, "Updating Runtime Overlays...\n")
			return overlay.BuildRuntimeOverlay(updateNodes)
		}
	}

	return nil
}
