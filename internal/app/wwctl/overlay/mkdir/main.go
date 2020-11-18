package mkdir

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/assets"
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
	"path"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	config := config.New()
	var overlaySourceDir string
//	mode := uint32(strconv.ParseUint(PermMode, 8, 32))


	if SystemOverlay == true {
		overlaySourceDir = config.SystemOverlaySource(args[0])
	} else {
		overlaySourceDir = config.RuntimeOverlaySource(args[0])
	}

	if util.IsDir(overlaySourceDir) == false {
		wwlog.Printf(wwlog.ERROR, "Overlay does not exist: %s\n", args[0])
		os.Exit(1)
	}

	overlayDir := path.Join(overlaySourceDir, args[1])

	wwlog.Printf(wwlog.DEBUG, "Will create directory in overlay: %s:%s\n", args[0], overlayDir)

	err := os.MkdirAll(overlayDir, 0755)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not create directory: %s\n", path.Dir(overlayDir))
		os.Exit(1)
	}

	fmt.Printf("Created directory within overlay: %s:%s\n", args[0], args[1])

	// Everything below this point is to update the relevant overlays
	nodes, err := assets.FindAllNodes()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Cloud not get nodeList: %s\n", err)
		os.Exit(1)
	}

	var updateNodes []assets.NodeInfo

	for _, node := range nodes {
		if SystemOverlay == true && node.SystemOverlay == args[0] {
			updateNodes = append(updateNodes, node)
		} else if node.RuntimeOverlay == args[0] {
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

	return nil
}