package chmod

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "chmod [OPTIONS] {runtime|system} OVERLAY_NAME FILENAME MODE",
		Short:                 "Change file permissions in an overlay",
		Long: `Changes the permissions of a single FILENAME within an overlay specified by
overlay type (system or runtime) and its OVERLAY_NAME.

You can use any MODE format supported by the chmod command.`,
		Example: "wwctl overlay chmod system default /etc/hostname.ww 0660",
		RunE:    CobraRunE,
		Args:    cobra.ExactArgs(4),
	}
)

func init() {
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
