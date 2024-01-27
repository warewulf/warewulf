package reload

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	return errors.Wrap(warewulfd.DaemonReload(), "failed to reload Warewulf server")
}
