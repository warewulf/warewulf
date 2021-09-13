package imprt

import (
	"os"
	"path"

	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var dest string
	var overlaySource string

	overlayKind := args[0]
	overlayName := args[1]
	source := args[2]

	if overlayKind != "system" && overlayKind != "runtime" {
		return errors.New("overlay kind must be of type 'system' or 'runtime'")
	}

	if len(args) == 4 {
		dest = args[3]
	} else {
		dest = source
	}

	if overlayKind == "system" {
		wwlog.Printf(wwlog.VERBOSE, "Copying '%s' into system overlay '%s:%s'\n", source, overlayName, dest)
		overlaySource = config.SystemOverlaySource(overlayName)
	} else if overlayKind == "runtime" {
		wwlog.Printf(wwlog.VERBOSE, "Copying '%s' into runtime overlay '%s:%s'\n", source, overlayName, dest)
		overlaySource = config.RuntimeOverlaySource(overlayName)
	}

	if !util.IsDir(overlaySource) {
		wwlog.Printf(wwlog.ERROR, "Overlay does not exist: %s:%s\n", overlayKind, overlayName)
		os.Exit(1)
	}

	err := util.CopyFile(source, path.Join(overlaySource, dest))
	if err != nil {
		return errors.Wrap(err, "could not copy file into overlay")
	}

	if !NoOverlayUpdate {
		n, err := node.New()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
			os.Exit(1)
		}

		nodes, err := n.FindAllNodes()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not get node list: %s\n", err)
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
