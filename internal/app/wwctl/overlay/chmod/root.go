package chmod

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:   "chmod [flags] <overlay kind> <overlay name> <path> <mode>",
		Short: "Change file permissions within an overlay",
		Long: "This command will allow you to change the permissions of a file within an\n" +
			"overlay.",
		RunE: CobraRunE,
		Args: cobra.ExactArgs(4),
	}
	NoOverlayUpdate bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&NoOverlayUpdate, "noupdate", "n", false, "Don't update overlays")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
