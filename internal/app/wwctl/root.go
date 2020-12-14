package wwctl

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/configure"
	"github.com/hpcng/warewulf/internal/app/wwctl/container"
	"github.com/hpcng/warewulf/internal/app/wwctl/kernel"
	"github.com/hpcng/warewulf/internal/app/wwctl/node"
	"github.com/hpcng/warewulf/internal/app/wwctl/overlay"
	"github.com/hpcng/warewulf/internal/app/wwctl/profile"
	"github.com/hpcng/warewulf/internal/app/wwctl/ready"
	"github.com/hpcng/warewulf/internal/app/wwctl/server"

	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:               "wwctl",
		Short:             "Warewulf Control",
		Long:              "Control interface to the Cluster Warewulf Provisioning System.",
		PersistentPreRunE: rootPersistentPreRunE,
	}
	verboseArg bool
	debugArg   bool
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verboseArg, "verbose", "v", false, "Run with increased verbosity.")
	rootCmd.PersistentFlags().BoolVarP(&debugArg, "debug", "d", false, "Run with debugging messages enabled.")

	//rootCmd.AddCommand(build.GetCommand())

	rootCmd.AddCommand(overlay.GetCommand())
	//	rootCmd.AddCommand(controller.GetCommand())
	rootCmd.AddCommand(container.GetCommand())
	rootCmd.AddCommand(node.GetCommand())
	rootCmd.AddCommand(kernel.GetCommand())
	//	rootCmd.AddCommand(group.GetCommand())
	rootCmd.AddCommand(profile.GetCommand())
	rootCmd.AddCommand(configure.GetCommand())
	rootCmd.AddCommand(ready.GetCommand())
	rootCmd.AddCommand(server.GetCommand())

}

// GetRootCommand returns the root cobra.Command for the application.
func GetRootCommand() *cobra.Command {
	return rootCmd
}

func rootPersistentPreRunE(cmd *cobra.Command, args []string) error {
	if verboseArg == true {
		wwlog.SetLevel(wwlog.VERBOSE)
	} else if debugArg == true {
		wwlog.SetLevel(wwlog.DEBUG)
	} else {
		wwlog.SetLevel(wwlog.INFO)
	}
	return nil
}
