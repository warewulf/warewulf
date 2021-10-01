package start

import (
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	if SetForeground {
		conf, err := warewulfconf.New()
		if err != nil {
			return errors.Wrap(err, "Could not read Warewulf configuration file")
		}
		conf.Warewulf.Syslog = false
		return errors.Wrap(warewulfd.RunServer(), "failed to start Warewulf server")
	} else {
		return errors.Wrap(warewulfd.DaemonStart(), "failed to start Warewulf server")
	}
}
