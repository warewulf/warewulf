package restart

import (
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	warewulfd.DaemonStop()
	return warewulfd.DaemonStart()
}
