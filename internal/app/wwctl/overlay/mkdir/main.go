package mkdir

import (
	"os"
	"path"

	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var overlaySourceDir string

	overlayKind := args[0]
	overlayName := args[1]
	dirName := args[2]

	if overlayKind != "system" && overlayKind != "runtime" {
		return errors.New("overlay kind must be of type 'system' or 'runtime'")
	}

	if overlayKind == "system" {
		overlaySourceDir = config.SystemOverlaySource(overlayName)
	} else if overlayKind == "runtime" {
		overlaySourceDir = config.RuntimeOverlaySource(overlayName)
	}

	if !util.IsDir(overlaySourceDir) {
		wwlog.Printf(wwlog.ERROR, "Overlay does not exist: %s:%s\n", overlayKind, overlayName)
		os.Exit(1)
	}

	overlayDir := path.Join(overlaySourceDir, dirName)

	wwlog.Printf(wwlog.DEBUG, "Will create directory in overlay: %s:%s:%s\n", overlayKind, overlayName, dirName)

	err := os.MkdirAll(overlayDir, os.FileMode(PermMode))
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not create directory: %s\n", path.Dir(overlayDir))
		os.Exit(1)
	}

	return nil
}
