package start

import (
	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	if SetForeground {
		conf := warewulfconf.New()
		conf.Warewulf.Syslog = false
		return errors.Wrap(warewulfd.RunServer(), "failed to start Warewulf server")
	} else {
		return errors.Wrap(warewulfd.DaemonStart(), "failed to start Warewulf server")
	}
}
