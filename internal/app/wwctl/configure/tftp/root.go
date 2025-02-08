package tftp

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "tftp [OPTIONS]",
		Short:                 "Manage and initialize TFTP",
		Long: "TFTP is a dependent service of Warewulf, this tool will enable the tftp services\n" +
			"on your Warewulf master.",
		RunE: CobraRunE,
		Args: cobra.ExactArgs(0),
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
