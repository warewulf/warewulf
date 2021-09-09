package chown

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
	fileName := args[1]
	var uid int
	var gid int
	var err error

	uid, err = strconv.Atoi(args[2])
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "UID is not an integer: %s\n", args[2])
		os.Exit(1)
	}

	if len(args) > 3 {
		gid, err = strconv.Atoi(args[3])
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "GID is not an integer: %s\n", args[3])
			os.Exit(1)
		}
	} else {
		gid = 0
	}

	if SystemOverlay {
		overlaySourceDir = config.SystemOverlaySource(overlayName)
	} else {
		overlaySourceDir = config.RuntimeOverlaySource(overlayName)
	}

	if !util.IsDir(overlaySourceDir) {
		wwlog.Printf(wwlog.ERROR, "Overlay does not exist: %s\n", overlayName)
		os.Exit(1)
	}

	overlayFile := path.Join(overlaySourceDir, fileName)

	if !util.IsFile(overlayFile) && !util.IsDir(overlayFile) {
		wwlog.Printf(wwlog.ERROR, "File does not exist within overlay: %s:%s\n", overlayName, fileName)
		os.Exit(1)
	}

	err = os.Chown(overlayFile, uid, gid)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not set ownership: %s\n", err)
		os.Exit(1)
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
			if SystemOverlay && node.SystemOverlay.Get() == overlayName {
				updateNodes = append(updateNodes, node)
			} else if node.RuntimeOverlay.Get() == overlayName {
				updateNodes = append(updateNodes, node)
			}
		}

		if SystemOverlay {
			wwlog.Printf(wwlog.INFO, "Updating System Overlays...\n")
			return overlay.BuildSystemOverlay(updateNodes)
		} else {
			wwlog.Printf(wwlog.INFO, "Updating Runtime Overlays...\n")
			return overlay.BuildRuntimeOverlay(updateNodes)
		}
	}

	return nil
}
