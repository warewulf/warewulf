package create

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	err = overlay.OverlayInit(args[0])
	return
}
