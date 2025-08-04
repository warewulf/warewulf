package mkdir

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {

	myOverlay, err := overlay.GetOverlay(args[0])
	if err != nil {
		return err
	}
	return myOverlay.Mkdir(args[1], PermMode)
}
