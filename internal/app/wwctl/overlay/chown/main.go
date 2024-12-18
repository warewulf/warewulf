package chown

import (
	"fmt"
	"os"
	"strconv"

	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/util"

	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var uid int
	var gid int
	var err error

	overlayName := args[0]
	fileName := args[1]

	uid, err = strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("UID is not an integer: %s", args[2])
	}

	if len(args) > 3 {
		gid, err = strconv.Atoi(args[3])
		if err != nil {
			return fmt.Errorf("GID is not an integer: %s", args[3])
		}
	} else {
		gid = -1
	}

	overlay_ := overlay.GetOverlay(overlayName)
	if !overlay_.Exists() {
		return fmt.Errorf("overlay does not exist: %s", overlayName)
	}

	if !overlay_.IsSiteOverlay() {
		overlay_, err = overlay_.CloneSiteOverlay()
		if err != nil {
			return err
		}
	}

	overlayFile := overlay_.File(fileName)
	if !(util.IsFile(overlayFile) || util.IsDir(overlayFile)) {
		return fmt.Errorf("file does not exist within overlay: %s:%s", overlayName, fileName)
	}

	err = os.Chown(overlayFile, uid, gid)
	if err != nil {
		return fmt.Errorf("could not set ownership: %s", err)
	}

	return nil
}
