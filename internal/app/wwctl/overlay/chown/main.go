package chown

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"strconv"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	fileName := args[1]

	uid, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("UID is not an integer: %s", args[2])
	}

	gid := -1
	if len(args) > 3 {
		gid, err = strconv.Atoi(args[3])
		if err != nil {
			return fmt.Errorf("GID is not an integer: %s", args[3])
		}
	}

	myOverlay, err := overlay.GetOverlay(args[0])
	if err != nil {
		return err
	}
	return myOverlay.Chown(fileName, uid, gid)
}
