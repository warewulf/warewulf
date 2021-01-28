package hosts

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:   "hosts",
		Short: "Update the /etc/hosts file",
		Long: "Write out the /etc/hosts file based on the Warewulf template (hosts.tmpl) in the\n" +
			"Warewulf configuration directory.",
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
