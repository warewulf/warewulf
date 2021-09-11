package start

import (
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	if SetForeground {
		return errors.Wrap(warewulfd.RunServer(), "failed to start Warewulf server")
	} else {
		return errors.Wrap(warewulfd.DaemonStart(), "failed to start Warewulf server")
	}
}
