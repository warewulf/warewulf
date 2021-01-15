package tftp

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:   "tftp",
		Short: "Manage and initialize TFTP",
		Long: "TFTP is a dependent service of Warewulf, this tool will enable the tftp services\n" +
			"on your Warewulf master.",
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
