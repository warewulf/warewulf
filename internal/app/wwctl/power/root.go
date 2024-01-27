package power

import (
	"github.com/spf13/cobra"
	powercycle "github.com/warewulf/warewulf/internal/app/wwctl/power/cycle"
	poweroff "github.com/warewulf/warewulf/internal/app/wwctl/power/off"
	poweron "github.com/warewulf/warewulf/internal/app/wwctl/power/on"
	powerreset "github.com/warewulf/warewulf/internal/app/wwctl/power/reset"
	powersoft "github.com/warewulf/warewulf/internal/app/wwctl/power/soft"
	powerstatus "github.com/warewulf/warewulf/internal/app/wwctl/power/status"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "power COMMAND [OPTIONS]",
		Short:                 "Warewulf node power management",
		Long:                  "This command controls the power state of nodes.",
	}
)

func init() {
	//	baseCmd.PersistentFlags().BoolVarP(&test, "test", "t", false, "Testing.")

	baseCmd.AddCommand(powercycle.GetCommand())
	baseCmd.AddCommand(poweroff.GetCommand())
	baseCmd.AddCommand(poweron.GetCommand())
	baseCmd.AddCommand(powerreset.GetCommand())
	baseCmd.AddCommand(powersoft.GetCommand())
	baseCmd.AddCommand(powerstatus.GetCommand())
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
