package tftp

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:   "tftp",
		Short: "TFTP configuration",
		Long:  "TFTP Config",
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
