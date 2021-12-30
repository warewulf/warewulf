package nodestatus

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "status [OPTIONS] [NODENAME...]",
		Short:                 "View the provisioning status of nodes",
		Long:                  "View and monitor the status of nodes as they are provisioned and check in.",
		RunE:                  CobraRunE,
		Args:                  cobra.MinimumNArgs(0),
	}
	SetWatch bool
)

func init() {
	baseCmd.PersistentFlags().BoolVar(&SetWatch, "watch", false, "Watch the status automatically")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
