package chown

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
	var uid int
	var gid int
	var err error

	overlayName := args[0]
	fileName := args[1]

	uid, err = strconv.Atoi(args[2])
	if err != nil {
		wwlog.Error("UID is not an integer: %s", args[2])
		os.Exit(1)
	}

	if len(args) > 3 {
		gid, err = strconv.Atoi(args[3])
		if err != nil {
			wwlog.Error("GID is not an integer: %s", args[3])
			os.Exit(1)
		}
	} else {
		gid = -1
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

	err = os.Chown(overlayFile, uid, gid)
	if err != nil {
		wwlog.Error("Could not set ownership: %s", err)
		os.Exit(1)
	}

	return nil
}
