package restart

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	err := warewulfd.DaemonStop()
	if err != nil {
		return errors.Wrap(err, "failed to stop Warewulf server")
	}
	return errors.Wrap(warewulfd.DaemonStart(), "failed to start Warewulf server")
}
