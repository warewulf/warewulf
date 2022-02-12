package hosts

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "hosts [OPTIONS]",
		Short:                 "Update the /etc/hosts file",
		Long: "Write out the /etc/hosts file based on the Warewulf template (hosts.tmpl) in the\n" +
			"Warewulf configuration directory.",
		RunE: CobraRunE,
	}
	setShow bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&setShow, "show", "s", false, "Show configuration (don't update)")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
