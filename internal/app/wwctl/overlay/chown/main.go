package chown

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	fileName := args[1]
	chownSpec := args[2]

	var uid, gid = -1, -1
	var err error

	if strings.Contains(chownSpec, ":") {
		parts := strings.SplitN(chownSpec, ":", 2)
		if parts[0] != "" {
			uid, err = strconv.Atoi(parts[0])
			if err != nil {
				return fmt.Errorf("UID is not an integer: %s", parts[0])
			}
		}
		if parts[1] != "" {
			gid, err = strconv.Atoi(parts[1])
			if err != nil {
				return fmt.Errorf("GID is not an integer: %s", parts[1])
			}
		}
	} else {
		uid, err = strconv.Atoi(chownSpec)
		if err != nil {
			return fmt.Errorf("UID is not an integer: %s", chownSpec)
		}
	}

	myOverlay, err := overlay.Get(args[0])
	if err != nil {
		return err
	}
	return myOverlay.Chown(fileName, uid, gid)
}
