package pull

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:   "pull",
		Short: "Pull Source OCI VNFS images",
		Long:  "Pull Source OCI VNFS images ",
		RunE:  CobraRunE,
		Args:  cobra.MinimumNArgs(1),
	}
	SetForce  bool
	SetUpdate bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&SetForce, "force", "f", false, "Force overwrite of an existing VNFS")
	baseCmd.PersistentFlags().BoolVarP(&SetUpdate, "update", "u", false, "Update and overwrite an existing VNFS")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
