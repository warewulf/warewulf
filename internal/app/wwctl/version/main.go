package version

import (
	"fmt"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/version"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	fmt.Println("wwctl version:\t", version.GetVersion())

	var wwVersionResponse *wwapiv1.VersionResponse = version.Version()
	fmt.Println("rpc version:", wwVersionResponse.String())
	return nil
}
