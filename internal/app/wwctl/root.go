package wwctl

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/configure"
	"github.com/hpcng/warewulf/internal/app/wwctl/container"
	"github.com/hpcng/warewulf/internal/app/wwctl/node"
	"github.com/hpcng/warewulf/internal/app/wwctl/overlay"
	"github.com/hpcng/warewulf/internal/app/wwctl/power"
	"github.com/hpcng/warewulf/internal/app/wwctl/profile"
	"github.com/hpcng/warewulf/internal/app/wwctl/server"
	"github.com/hpcng/warewulf/internal/app/wwctl/version"
	"github.com/hpcng/warewulf/internal/pkg/help"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"io"
)

var (
	rootCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "wwctl COMMAND [OPTIONS]",
		Short:                 "Warewulf Control",
		Long:                  "Control interface to the Warewulf Cluster Provisioning System.",
		PersistentPreRunE:     rootPersistentPreRunE,
		SilenceUsage:          true,
		SilenceErrors:         true,
	}
	verboseArg bool
	DebugFlag  bool
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verboseArg, "verbose", "v", false, "Run with increased verbosity.")
	rootCmd.PersistentFlags().BoolVarP(&DebugFlag, "debug", "d", false, "Run with debugging messages enabled.")

	rootCmd.SetUsageTemplate(help.UsageTemplate)
	rootCmd.SetHelpTemplate(help.HelpTemplate)

	rootCmd.AddCommand(overlay.GetCommand())
	rootCmd.AddCommand(container.GetCommand())
	rootCmd.AddCommand(node.GetCommand())
	rootCmd.AddCommand(power.GetCommand())
	rootCmd.AddCommand(profile.GetCommand())
	rootCmd.AddCommand(configure.GetCommand())
	rootCmd.AddCommand(server.GetCommand())
	rootCmd.AddCommand(version.GetCommand())

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

// External functions not used by the wwctl command line

// Generate Bash completion file
func GenBashCompletion(w io.Writer) error {
	return rootCmd.GenBashCompletion(w)
}

// Generate man pages
func GenManTree(fileName string) error {
	header := &doc.GenManHeader{
		Title:   "WWCTL",
		Section: "1",
	}
	return doc.GenManTree(rootCmd, header, fileName)
}
