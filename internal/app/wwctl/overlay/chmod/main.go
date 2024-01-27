package chmod

import (
	"os"
	"path"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var overlaySourceDir string

	overlayName := args[0]
	fileName := args[1]

	permissionMode, err := strconv.ParseUint(args[2], 8, 32)
	if err != nil {
		wwlog.Error("Could not convert requested mode: %s", err)
		os.Exit(1)
	}

	overlaySourceDir = overlay.OverlaySourceDir(overlayName)

	if !util.IsDir(overlaySourceDir) {
		wwlog.Error("Overlay does not exist: %s", overlayName)
		os.Exit(1)
	}

	overlayFile := path.Join(overlaySourceDir, fileName)

	if !util.IsFile(overlayFile) && !util.IsDir(overlayFile) {
		wwlog.Error("File does not exist within overlay: %s:%s", overlayName, fileName)
		os.Exit(1)
	}

	err = os.Chmod(overlayFile, os.FileMode(permissionMode))
	if err != nil {
		wwlog.Error("Could not set permission: %s", err)
		os.Exit(1)
	}

	return nil
}
