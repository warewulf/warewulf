package chmod

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	myOverlay, err := overlay.Get(args[0])
	if err != nil {
		return err
	}
	path := args[1]

	permissionMode, err := strconv.ParseUint(args[2], 8, 32)
	if err != nil {
		return fmt.Errorf("could not convert requested mode: %s", err)
	}
	return myOverlay.Chmod(path, permissionMode)
}
