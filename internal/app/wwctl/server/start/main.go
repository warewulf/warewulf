package start

import (
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	if SetForeground {
		return warewulfd.RunServer()
	} else {
		return warewulfd.DaemonStart()
	}
}
