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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var overlaySourceDir string

	overlayKind := args[0]
	overlayName := args[1]
	fileName := args[2]

	if overlayKind != "system" && overlayKind != "runtime" {
		return errors.New("overlay kind must be of type 'system' or 'runtime'")
	}

	permissionMode, err := strconv.ParseInt(args[3], 8, 32)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not convert requested mode: %s\n", err)
		os.Exit(1)
	}

	if overlayKind == "system" {
		overlaySourceDir = config.SystemOverlaySource(overlayName)
	} else if overlayKind == "runtime" {
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

	err = os.Chmod(overlayFile, os.FileMode(permissionMode))
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not set permission: %s\n", err)
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
