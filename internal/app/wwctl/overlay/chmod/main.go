package chmod

import (
	"os"
	"path"
	"strconv"

	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var overlaySourceDir string

	overlayName := args[0]
	fileName := args[1]

	permissionMode, err := strconv.ParseUint(args[2], 8, 32)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not convert requested mode: %s\n", err)
		os.Exit(1)
	}

	overlaySourceDir = overlay.OverlaySourceDir(overlayName)

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

	return nil
}
