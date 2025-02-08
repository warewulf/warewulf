package nodestatus

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "status [OPTIONS] [NODENAME...]",
		Short:                 "View the provisioning status of nodes",
		Long:                  "View and monitor the status of nodes as they are provisioned and check in.",
		RunE:                  CobraRunE,
		ValidArgsFunction:     completions.Nodes(0), // no limit
	}
	SetWatch       bool
	SetUpdate      int
	SetTime        int64
	SetSortLast    bool
	SetSortReverse bool
	SetUnknown     bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&SetWatch, "watch", "w", false, "Watch the status automatically")
	baseCmd.PersistentFlags().IntVarP(&SetUpdate, "update", "U", 500, "Set the update frequency for 'watch' (ms)")
	baseCmd.PersistentFlags().Int64VarP(&SetTime, "time", "t", 0, "Filter by last checkin time (seconds)")
	baseCmd.PersistentFlags().BoolVarP(&SetSortLast, "last", "l", false, "Sort by the last check-in time")
	baseCmd.PersistentFlags().BoolVarP(&SetSortReverse, "reverse", "r", false, "Reverse the sort order")
	baseCmd.PersistentFlags().BoolVarP(&SetUnknown, "unknown", "u", false, "Only show nodes of unknown status")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
