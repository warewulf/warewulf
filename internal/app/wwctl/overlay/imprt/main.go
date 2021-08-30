package imprt

import (
	"os"
	"path"

	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	overlayName := args[0]
	source := args[1]
	var dest string
	var overlaySource string

	if len(args) == 3 {
		dest = args[2]
	} else {
		dest = source
	}

	if SystemOverlay == true {
		wwlog.Printf(wwlog.VERBOSE, "Importing '%s' into system overlay '%s:%s'\n", source, overlayName, dest)
		overlaySource = config.SystemOverlaySource(overlayName)
	} else {
		wwlog.Printf(wwlog.VERBOSE, "Importing '%s' into runtime overlay '%s:%s'\n", source, overlayName, dest)
		overlaySource = config.RuntimeOverlaySource(overlayName)
	}

	if util.IsDir(overlaySource) == false {
		wwlog.Printf(wwlog.ERROR, "Overlay does not exist: %s\n", overlayName)
		os.Exit(1)
	}

	err := util.CopyFile(source, path.Join(overlaySource, dest))
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Failed copying file into overlay sourcedir:\n")
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	if NoOverlayUpdate == false {
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
			if SystemOverlay == true && node.SystemOverlay.Get() == overlayName {
				updateNodes = append(updateNodes, node)
			} else if node.RuntimeOverlay.Get() == overlayName {
				updateNodes = append(updateNodes, node)
			}
		}

		if SystemOverlay == true {
			wwlog.Printf(wwlog.INFO, "Updating System Overlays...\n")
			return overlay.BuildSystemOverlay(updateNodes)
		} else {
			wwlog.Printf(wwlog.INFO, "Updating Runtime Overlays...\n")
			return overlay.BuildRuntimeOverlay(updateNodes)
		}

	}

	return nil
}
