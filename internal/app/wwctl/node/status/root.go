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
	SetWatch  bool
	SetUpdate int
	SetTime  int64
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&SetWatch, "watch", "w", false, "Watch the status automatically")
	baseCmd.PersistentFlags().IntVarP(&SetUpdate, "update", "u", 500, "Set the update frequency for 'watch' (ms)")
	baseCmd.PersistentFlags().Int64VarP(&SetTime, "time", "t", 0, "Filter by last checkin time (seconds)")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
