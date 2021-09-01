package chmod

import (
	"os"
	"path"
	"strconv"

	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var overlaySourceDir string
	overlayName := args[0]
	fileName := args[2]

	permissionMode, err := strconv.ParseInt(args[1], 8, 32)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not convert requested mode: %s\n", err)
		os.Exit(1)
	}

	if SystemOverlay == true {
		overlaySourceDir = config.SystemOverlaySource(overlayName)
	} else {
		overlaySourceDir = config.RuntimeOverlaySource(overlayName)
	}

	if util.IsDir(overlaySourceDir) == false {
		wwlog.Printf(wwlog.ERROR, "Overlay does not exist: %s\n", overlayName)
		os.Exit(1)
	}

	overlayFile := path.Join(overlaySourceDir, fileName)

	if util.IsFile(overlayFile) == false {
		wwlog.Printf(wwlog.ERROR, "File does not exist within overlay: %s:%s\n", overlayName, fileName)
		os.Exit(1)
	}

	err = os.Chmod(overlayFile, os.FileMode(permissionMode))
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not set permission: %s\n", err)
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
