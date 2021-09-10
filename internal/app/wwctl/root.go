package wwctl

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/configure"
	"github.com/hpcng/warewulf/internal/app/wwctl/container"
	"github.com/hpcng/warewulf/internal/app/wwctl/kernel"
	"github.com/hpcng/warewulf/internal/app/wwctl/node"
	"github.com/hpcng/warewulf/internal/app/wwctl/overlay"
	"github.com/hpcng/warewulf/internal/app/wwctl/power"
	"github.com/hpcng/warewulf/internal/app/wwctl/profile"
	"github.com/hpcng/warewulf/internal/app/wwctl/server"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"io"
)

var (
	rootCmd = &cobra.Command{
		Use:               "wwctl",
		Short:             "Warewulf Control",
		Long:              "Control interface to the Cluster Warewulf Provisioning System.",
		PersistentPreRunE: rootPersistentPreRunE,
		SilenceUsage:      true,
	}
	verboseArg bool
	DebugFlag  bool
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verboseArg, "verbose", "v", false, "Run with increased verbosity.")
	rootCmd.PersistentFlags().BoolVarP(&DebugFlag, "debug", "d", false, "Run with debugging messages enabled.")

	rootCmd.AddCommand(overlay.GetCommand())
	rootCmd.AddCommand(container.GetCommand())
	rootCmd.AddCommand(node.GetCommand())
	rootCmd.AddCommand(kernel.GetCommand())
	rootCmd.AddCommand(power.GetCommand())
	rootCmd.AddCommand(profile.GetCommand())
	rootCmd.AddCommand(configure.GetCommand())
	rootCmd.AddCommand(server.GetCommand())

}

// GetRootCommand returns the root cobra.Command for the application.
func GetRootCommand() *cobra.Command {
	return rootCmd
}

func rootPersistentPreRunE(cmd *cobra.Command, args []string) error {
	if DebugFlag {
		wwlog.SetLevel(wwlog.DEBUG)
	} else if verboseArg {
		wwlog.SetLevel(wwlog.VERBOSE)
	} else {
		wwlog.SetLevel(wwlog.INFO)
	}
	return nil
}

// GenBashCompletionFile
func GenBashCompletion(w io.Writer) error {
	return rootCmd.GenBashCompletion(w)
}

func GenManTree(fileName string) error {
	header := &doc.GenManHeader{
		Title:   "MINE",
		Section: "1",
	}
	return doc.GenManTree(rootCmd, header, fileName)
}
