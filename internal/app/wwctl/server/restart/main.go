package restart

import (
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	err := warewulfd.DaemonStop()
	if err != nil {
		return errors.Wrap(err, "failed to stop Warewulf server")
	}
	return errors.Wrap(warewulfd.DaemonStart(), "failed to start Warewulf server")
}
