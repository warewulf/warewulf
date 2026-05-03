package overlaydiff

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwclient/overlaydiff/capture"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "overlay-diff COMMAND [OPTIONS]",
		Short:                 "Inspect differences between source and baseline trees",
		Long:                  "Inspect differences between source and baseline trees for overlay authoring",
		Args:                  cobra.NoArgs,
	}
)

func init() {
	baseCmd.AddCommand(capture.GetStartCommand())
	baseCmd.AddCommand(capture.GetStopCommand())
}

func GetCommand() *cobra.Command {
	return baseCmd
}
