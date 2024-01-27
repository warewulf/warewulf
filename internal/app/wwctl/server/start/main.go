package start

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	if SetForeground {
		conf := warewulfconf.Get()
		conf.Warewulf.Syslog = false
		return errors.Wrap(warewulfd.RunServer(), "failed to start Warewulf server")
	} else {
		return errors.Wrap(warewulfd.DaemonStart(), "failed to start Warewulf server")
	}
}
