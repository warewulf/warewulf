package sensors

import (
	"github.com/spf13/cobra"
)

var (
	powerCmd = &cobra.Command{
		Use:   "sensors",
		Short: "node sensor information",
		Long:  "Get Node Sensor Information",
		RunE:  CobraRunE,
	}
	full bool
)

func init() {
        powerCmd.PersistentFlags().BoolVarP(&full, "full", "F", false, "show detailed output.")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return powerCmd
}
