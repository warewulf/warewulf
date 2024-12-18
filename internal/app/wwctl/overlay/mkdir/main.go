package mkdir

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	overlayName := args[0]
	dirName := args[1]

	overlay_ := overlay.GetOverlay(overlayName)
	if !overlay_.IsSiteOverlay() {
		overlay_, err = overlay_.CloneSiteOverlay()
		if err != nil {
			return err
		}
	}

	if !overlay_.Exists() {
		return fmt.Errorf("overlay does not exist: %s", overlayName)
	}

	overlayDir := overlay_.File(dirName)
	wwlog.Debug("Will create directory in overlay: %s:%s", overlayName, dirName)
	err = os.MkdirAll(overlayDir, os.FileMode(PermMode))
	if err != nil {
		return fmt.Errorf("could not create directory: %s", overlayDir)
	}

	return nil
}
