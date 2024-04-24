package rename

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/container"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	if len(args) != 2 {
		return fmt.Errorf("rename requires 2 arguments: %d provided", len(args))
	}

	err = container.Rename(&container.RenameParameter{
		Name:       args[0],
		TargetName: args[1],
		Build:      SetBuild,
	})

	if err != nil {
		return
	}
	wwlog.Info("Container %s successfully renamed as %s", args[0], args[1])

	err = warewulfd.DaemonStatus()
	if err != nil {
		// warewulfd is not running, skip
		return
	}
	// else reload daemon to apply new changes
	return warewulfd.DaemonReload()
}
