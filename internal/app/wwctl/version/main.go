package version

import (
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/version"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	wwlog.Info("wwctl version:\t", version.GetVersion())

	var wwVersionResponse *wwapiv1.VersionResponse = version.Version()
	wwlog.Info("rpc version:", wwVersionResponse.String())
	return nil
}
