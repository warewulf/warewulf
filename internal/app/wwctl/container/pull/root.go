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
	SetForce   bool
	SetUpdate  bool
	SetBuild   bool
	SetDefault bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&SetForce, "force", "f", false, "Force overwrite of an existing container")
	baseCmd.PersistentFlags().BoolVarP(&SetUpdate, "update", "u", false, "Update and overwrite an existing container")
	baseCmd.PersistentFlags().BoolVarP(&SetBuild, "build", "b", false, "Build container when after pulling")
	baseCmd.PersistentFlags().BoolVar(&SetDefault, "setdefault", false, "Set this container for the default profile")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
