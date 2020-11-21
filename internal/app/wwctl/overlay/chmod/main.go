package chmod

import (
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
	"path"
	"strconv"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	config := config.New()
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
			wwlog.Printf(wwlog.ERROR, "Cloud not get nodeList: %s\n", err)
			os.Exit(1)
		}

		var updateNodes []node.NodeInfo

		for _, node := range nodes {
			if SystemOverlay == true && node.SystemOverlay == overlayName {
				updateNodes = append(updateNodes, node)
			} else if node.RuntimeOverlay == overlayName {
				updateNodes = append(updateNodes, node)
			}
		}

		if SystemOverlay == true {
			wwlog.Printf(wwlog.INFO, "Updating System Overlays...\n")
			return overlay.SystemBuild(updateNodes, true)
		} else {
			wwlog.Printf(wwlog.INFO, "Updating Runtime Overlays...\n")
			return overlay.RuntimeBuild(updateNodes, true)
		}
	}

	return nil
}