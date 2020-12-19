package nfs

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:   "nfs",
		Short: "NFS configuration",
		Long:  "NFS Config",
		RunE:  CobraRunE,
	}
	SetShow bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&SetShow, "show", "s", false, "Show configuration (don't update)")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
