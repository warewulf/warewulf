package ssh

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:   "ssh",
		Short: "Manage and initialize SSH",
		Long: "SSH is an optionally dependent service for Warewulf, this tool will automatically\n" +
			"setup the ssh keys nodes using the 'default' system overlay as well as user owned\n" +
			"keys.",
		RunE: CobraRunE,
	}
	SetPersist bool
)

func init() {
	baseCmd.PersistentFlags().BoolVar(&SetPersist, "persist", false, "Persist the configuration and initialize the service")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
