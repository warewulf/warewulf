package exec

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/container/exec/child"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:   "exec [flags] [container name]",
		Short: "Run a command inside of a Warewulf container",
		Long: "This command will allow you to run any command inside of a given\n" +
			"warewulf container. This is commonly used with an interactive shell such as /bin/bash\n" +
			"to run a virtual environment within the container.",
		RunE:               CobraRunE,
		Args:               cobra.MinimumNArgs(1),
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	}
	binds []string
)

func init() {
	baseCmd.AddCommand(child.GetCommand())
	baseCmd.PersistentFlags().StringArrayVarP(&binds, "bind", "b", []string{}, "Bind a local path into the container (must exist)")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
