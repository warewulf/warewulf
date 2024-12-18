package chmod

import (
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/util"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var overlaySourceDir string

	overlayName := args[0]
	fileName := args[1]

	permissionMode, err := strconv.ParseUint(args[2], 8, 32)
	if err != nil {
		return fmt.Errorf("could not convert requested mode: %s", err)
	}
	err = overlay.CloneSiteOverlay(overlayName)
	if err != nil {
		return err
	}
	overlaySourceDir, _ = overlay.GetOverlay(overlayName)

	if !util.IsDir(overlaySourceDir) {
		return fmt.Errorf("overlay does not exist: %s", overlayName)
	}

	overlayFile := path.Join(overlaySourceDir, fileName)

	if !util.IsFile(overlayFile) && !util.IsDir(overlayFile) {
		return fmt.Errorf("file does not exist within overlay: %s:%s", overlayName, fileName)
	}

	err = os.Chmod(overlayFile, os.FileMode(permissionMode))
	if err != nil {
		return fmt.Errorf("could not set permission: %s", err)
	}

	return nil
}
