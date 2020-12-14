package exec

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/container/exec/child"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:                "exec",
		Short:              "Spawn any command inside a Warewulf container",
		Long:               "Run a command inside a Warewulf container ",
		RunE:               CobraRunE,
		Args:               cobra.MinimumNArgs(1),
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	}
)

func init() {
	baseCmd.AddCommand(child.GetCommand())

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
