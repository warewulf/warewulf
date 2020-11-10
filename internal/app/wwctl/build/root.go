package build

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	buildCmd = &cobra.Command{
		Use:                "build",
		Short:              "Warewulf build subcommand",
		Long:               "Warewulf build is used to build VNFS, kernel, and overlay objects for provisioning.",
		RunE:				buildRunE,
	}
	option string
)

func init() {
	//buildCmd.PersistentFlags().StringVar(&flagHost, FlagHost, "localhost:7331", "Address of the Fuzzball Host.")

	buildCmd.PersistentFlags().StringVarP(&option, "testopt", "t", "default", "This is a test option")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return buildCmd
}

func buildRunE(cmd *cobra.Command, args []string) error {
	fmt.Printf("Hello World. Got Option: %s, %s\n", option, args)

	return nil
}

