package imprt

import (
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
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
	overlaySource = overlay.OverlaySourceDir(overlayName)

	if !util.IsDir(overlaySource) {
		wwlog.Error("Overlay does not exist: %s", overlayName)
		os.Exit(1)
	}

	if util.IsDir(path.Join(overlaySource, dest)) {
		dest = path.Join(dest, path.Base(source))
	}

	if util.IsFile(path.Join(overlaySource, dest)) {
		wwlog.Error("A file with that name already exists in the overlay %s\n:", overlayName)
		os.Exit(1)
	}

	if CreateDirs {
		parent := filepath.Dir(path.Join(overlaySource, dest))
		if _, err := os.Stat(parent); os.IsNotExist(err) {
			wwlog.Debug("Create dir: %s", parent)
			srcInfo, err := os.Stat(source)
			if err != nil {
				wwlog.Error("Could not retrieve the stat for file: %s", err)
				return err
			}
			err = os.MkdirAll(parent, srcInfo.Mode())
			if err != nil {
				wwlog.Error("Could not create parent dif: %s: %v", parent, err)
				return err
			}
		}
	}

	err := util.CopyFile(source, path.Join(overlaySource, dest))
	if err != nil {
		return errors.Wrap(err, "could not copy file into overlay")
	}

	if !NoOverlayUpdate {
		n, err := node.New()
		if err != nil {
			wwlog.Error("Could not open node configuration: %s", err)
			os.Exit(1)
		}

		nodes, err := n.FindAllNodes()
		if err != nil {
			wwlog.Error("Could not get node list: %s", err)
			os.Exit(1)
		}

		var updateNodes []node.NodeConf

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
