package mkdir

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	var overlaySourceDir string

	overlayName := args[0]
	dirName := args[1]
	err = overlay.CreateSiteOverlay(overlayName)
	if err != nil {
		return err
	}
	overlaySourceDir, _ = overlay.OverlaySourceDir(overlayName)

	if !util.IsDir(overlaySourceDir) {
		return fmt.Errorf("overlay does not exist: %s", overlayName)
	}

	overlayDir := path.Join(overlaySourceDir, dirName)

	wwlog.Debug("Will create directory in overlay: %s:%s", overlayName, dirName)

	err = os.MkdirAll(overlayDir, os.FileMode(PermMode))
	if err != nil {
		return fmt.Errorf("could not create directory: %s", path.Dir(overlayDir))
	}

	return nil
}
