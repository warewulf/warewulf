package stop

import (
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	return errors.Wrap(warewulfd.DaemonStop(), "failed to stop Warewulf server")
}
