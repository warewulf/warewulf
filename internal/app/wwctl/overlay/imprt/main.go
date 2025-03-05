package imprt

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	var dest string

	overlayName := args[0]
	source := args[1]

	if len(args) == 3 {
		dest = args[2]
	} else {
		dest = source
	}

	wwlog.Verbose("Copying '%s' into overlay '%s:%s'", source, overlayName, dest)
	overlay_ := overlay.GetOverlay(overlayName)
	if !overlay_.IsSiteOverlay() {
		overlay_, err = overlay_.CloneSiteOverlay()
		if err != nil {
			return err
		}
	}
	if !overlay_.Exists() {
		return fmt.Errorf("overlay does not exist: %s", overlayName)
	}

	if util.IsDir(overlay_.File(dest)) {
		dest = path.Join(dest, path.Base(source))
	}

	if !OverwriteFile && util.IsFile(overlay_.File(dest)) {
		return fmt.Errorf("a file with that name already exists in the overlay: %s", overlayName)
	}

	if CreateDirs {
		parent := filepath.Dir(overlay_.File(dest))
		if _, err = os.Stat(parent); os.IsNotExist(err) {
			wwlog.Debug("Create dir: %s", parent)
			srcInfo, err := os.Stat(source)
			if err != nil {
				return fmt.Errorf("could not retrieve the stat for file: %w", err)
			}
			mode := srcInfo.Mode()
			mode |= ((mode & 0444) >> 2) // add execute permission wherever srcInfo has read
			err = os.MkdirAll(parent, mode)
			if err != nil {
				return fmt.Errorf("could not create parent dir: %s: %w", parent, err)
			}
		}
	}

	err = util.CopyFile(source, overlay_.File(dest))
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

		workers := Workers
		if workers <= 0 {
			workers = runtime.NumCPU()
		}
		return overlay.BuildSpecificOverlays(updateNodes, nodes, []string{overlayName}, workers)
	}

	return nil
}
