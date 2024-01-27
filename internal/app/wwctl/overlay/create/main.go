package create

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	err := overlay.OverlayInit(args[0])
	if err != nil {
		wwlog.Error("%s", err)
		os.Exit(1)
	}

	return nil
}
