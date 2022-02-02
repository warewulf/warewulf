package mkdir

import (
	"os"
	"path"

	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var overlaySourceDir string

	overlayName := args[0]
	dirName := args[1]

	overlaySourceDir = overlay.OverlaySourceDir(overlayName)

	if !util.IsDir(overlaySourceDir) {
		wwlog.Printf(wwlog.ERROR, "Overlay does not exist: %s\n", overlayName)
		os.Exit(1)
	}

	overlayDir := path.Join(overlaySourceDir, dirName)

	wwlog.Printf(wwlog.DEBUG, "Will create directory in overlay: %s:%s\n", overlayName, dirName)

	err := os.MkdirAll(overlayDir, os.FileMode(PermMode))
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not create directory: %s\n", path.Dir(overlayDir))
		os.Exit(1)
	}

	return nil
}
