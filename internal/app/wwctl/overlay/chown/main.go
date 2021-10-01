package chown

import (
	"os"
	"path"
	"strconv"

	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var overlaySourceDir string
	var uid int
	var gid int
	var err error

	overlayKind := args[0]
	overlayName := args[1]
	fileName := args[2]

	if overlayKind != "system" && overlayKind != "runtime" {
		return errors.New("overlay kind must be of type 'system' or 'runtime'")
	}

	uid, err = strconv.Atoi(args[3])
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "UID is not an integer: %s\n", args[3])
		os.Exit(1)
	}

	if len(args) > 4 {
		gid, err = strconv.Atoi(args[4])
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "GID is not an integer: %s\n", args[4])
			os.Exit(1)
		}
	} else {
		gid = 0
	}

	if overlayKind == "system" {
		overlaySourceDir = config.SystemOverlaySource(overlayName)
	} else if overlayKind == "runtime" {
		overlaySourceDir = config.RuntimeOverlaySource(overlayName)
	}

	if !util.IsDir(overlaySourceDir) {
		wwlog.Printf(wwlog.ERROR, "Overlay does not exist: %s\n", overlayName)
		os.Exit(1)
	}

	overlayFile := path.Join(overlaySourceDir, fileName)

	if !util.IsFile(overlayFile) && !util.IsDir(overlayFile) {
		wwlog.Printf(wwlog.ERROR, "File does not exist within overlay: %s:%s\n", overlayName, fileName)
		os.Exit(1)
	}

	err = os.Chown(overlayFile, uid, gid)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not set ownership: %s\n", err)
		os.Exit(1)
	}

	return nil
}
