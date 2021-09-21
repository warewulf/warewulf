package set

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:   "set [flags] [kernel version]",
		Short: "Configure kernel properties",
		Long: "This command will allow you to set configuration properties for kernels.\n\n" +
			"Note: use the string 'UNSET' to remove a configuration",
		Args: cobra.MinimumNArgs(1),
		RunE: CobraRunE,
	}
	SetDefault   bool
)

func init() {
	baseCmd.PersistentFlags().BoolVar(&SetDefault, "setdefault", false, "Set this kernel for the default profile")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
