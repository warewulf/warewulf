package nfs

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:   "nfs",
		Short: "Manage and initialize NFS",
		Long: "NFS is an optional dependent service of Warewulf, this tool will automatically\n" +
			"configure NFS as per the configuration in the warewulf.conf file.",
		RunE: CobraRunE,
	}
	SetShow    bool
	SetPersist bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&SetShow, "show", "s", false, "Show configuration (don't update)")
	baseCmd.PersistentFlags().BoolVar(&SetPersist, "persist", false, "Persist the configuration and initialize the service")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
