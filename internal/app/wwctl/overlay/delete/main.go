package delete

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	fileName := ""
	if len(args) == 2 {
		fileName = args[1]
	}

	myOverlay, err := overlay.Get(args[0])
	if err != nil {
		return err
	}

	if fileName == "" {
		return myOverlay.Delete(Force)
	} else {
		return myOverlay.DeleteFile(fileName, Force, Parents)
	}
}
