package imprt

import (
	"os"
	"path"

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

	overlayName := args[0]
	source := args[1]

	if len(args) == 3 {
		dest = args[2]
	} else {
		dest = source
	}

	wwlog.Printf(wwlog.VERBOSE, "Copying '%s' into overlay '%s:%s'\n", source, overlayName, dest)
	overlaySource = overlay.OverlaySourceDir(overlayName)

	if !util.IsDir(overlaySource) {
		wwlog.Printf(wwlog.ERROR, "Overlay does not exist: %s\n", overlayName)
		os.Exit(1)
	}

	if util.IsDir(path.Join(overlaySource, dest)) {
		dest = path.Join(dest, path.Base(source))
	}

	if util.IsFile(path.Join(overlaySource, dest)) {
		wwlog.Printf(wwlog.ERROR, "A file with that name already exists in the overlay %s\n:", overlayName)
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
			if node.SystemOverlay.Get() == overlayName {
				updateNodes = append(updateNodes, node)
			} else if node.RuntimeOverlay.Get() == overlayName {
				updateNodes = append(updateNodes, node)
			}
		}

		return overlay.BuildSpecificOverlays(updateNodes, overlayName)
	}

	return nil
}
