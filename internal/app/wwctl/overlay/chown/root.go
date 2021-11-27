package chown

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:   "chown [OPTIONS] {system|runtime} OVERLAY_NAME FILE UID [GID]",
		Short: "Change file ownership within an overlay",
		Long: "This command changes the ownership of a FILE within the system or runtime OVERLAY_NAME\n" +
			"to the user specified by UID. Optionally, it will also change group ownership to GID.",
		RunE: CobraRunE,
		Args: cobra.RangeArgs(4, 5),
	}
)

func init() {
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
