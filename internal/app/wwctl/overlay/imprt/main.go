package imprt

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var dest string
	var overlaySource string

	overlayName := args[0]
	source := args[1]

	if len(args) == 3 {
		dest = args[2]
	} else {
		dest = source
	}

	wwlog.Verbose("Copying '%s' into overlay '%s:%s'", source, overlayName, dest)
	overlaySource, _ = overlay.OverlaySourceDir(overlayName)

	if !util.IsDir(overlaySource) {
		return fmt.Errorf("overlay does not exist: %s", overlayName)
	}

	if util.IsDir(path.Join(overlaySource, dest)) {
		dest = path.Join(dest, path.Base(source))
	}

	if util.IsFile(path.Join(overlaySource, dest)) {
		return fmt.Errorf("a file with that name already exists in the overlay: %s", overlayName)
	}

	if CreateDirs {
		parent := filepath.Dir(path.Join(overlaySource, dest))
		if _, err := os.Stat(parent); os.IsNotExist(err) {
			wwlog.Debug("Create dir: %s", parent)
			srcInfo, err := os.Stat(source)
			if err != nil {
				return fmt.Errorf("could not retrieve the stat for file: %s", err)
			}
			err = os.MkdirAll(parent, srcInfo.Mode())
			if err != nil {
				return fmt.Errorf("could not create parent dif: %s: %v", parent, err)
			}
		}
	}

	err := util.CopyFile(source, path.Join(overlaySource, dest))
	if err != nil {
		return fmt.Errorf("could not copy file into overlay: %w", err)
	}

	if !NoOverlayUpdate {
		n, err := node.New()
		if err != nil {
			return fmt.Errorf("could not open node configuration: %s", err)
		}

		nodes, err := n.FindAllNodes()
		if err != nil {
			return fmt.Errorf("could not get node list: %s", err)
		}

		var updateNodes []node.Node

		for _, node := range nodes {
			if util.InSlice(node.SystemOverlay, overlayName) {
				updateNodes = append(updateNodes, node)
			} else if util.InSlice(node.RuntimeOverlay, overlayName) {
				updateNodes = append(updateNodes, node)
			}
		}

		return overlay.BuildSpecificOverlays(updateNodes, []string{overlayName})
	}

	return nil
}
