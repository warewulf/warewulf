package wwctl

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/build"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:                "wwctl",
		Short:              "Warewulf CTL",
		Long:               "Fuzzball CLI is an application for interacting with a Fuzzball Service.",
		PersistentPreRunE:  rootPersistentPreRunE,
	}
	verboseArg bool
	debugArg bool
)



func init() {
	rootCmd.PersistentFlags().BoolVarP(&verboseArg, "verbose", "v", false, "Run with increased verbosity.")
	rootCmd.PersistentFlags().BoolVarP(&debugArg, "debug", "d", false, "Run with debugging messages enabled.")

	rootCmd.AddCommand(build.GetCommand())
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