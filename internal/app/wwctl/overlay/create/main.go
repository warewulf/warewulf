package create

import (
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
)



func CobraRunE(cmd *cobra.Command, args []string) error {

	if SystemOverlay == true {
		err := overlay.SystemOverlayInit(args[0])
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}
		wwlog.Printf(wwlog.INFO, "Created new system overlay: %s\n", args[0])
	} else {
		err := overlay.RuntimeOverlayInit(args[0])
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}
		wwlog.Printf(wwlog.INFO, "Created new runtime overlay: %s\n", args[0])
	}

	return nil
}