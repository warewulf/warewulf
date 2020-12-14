package delete

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete Source OCI VNFS images",
		Long:  "Delete Source OCI VNFS images ",
		RunE:  CobraRunE,
		Args:  cobra.MinimumNArgs(1),
	}
	SetForce  bool
	SetUpdate bool
	SetBuild  bool
)

func init() {

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
