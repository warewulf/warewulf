package mkdir

import (
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var overlaySourceDir string

	overlayName := args[0]
	dirName := args[1]

	overlaySourceDir = overlay.OverlaySourceDir(overlayName)

	if !util.IsDir(overlaySourceDir) {
		wwlog.Error("Overlay does not exist: %s", overlayName)
		os.Exit(1)
	}

	overlayDir := path.Join(overlaySourceDir, dirName)

	wwlog.Debug("Will create directory in overlay: %s:%s", overlayName, dirName)

	err := os.MkdirAll(overlayDir, os.FileMode(PermMode))
	if err != nil {
		wwlog.Error("Could not create directory: %s", path.Dir(overlayDir))
		os.Exit(1)
	}

	return nil
}
