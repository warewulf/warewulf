package chmod

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "chmod [OPTIONS] OVERLAY_NAME FILENAME MODE",
		Short:                 "Change file permissions in an overlay",
		Long:                  "Changes the permissions of a single FILENAME within an overlay.\nYou can use any MODE format supported by the chmod command.",
		Example:               "wwctl overlay chmod default /etc/hostname.ww 0660",
		RunE:                  CobraRunE,
		Args:                  cobra.ExactArgs(3),
	}
)

func init() {
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
