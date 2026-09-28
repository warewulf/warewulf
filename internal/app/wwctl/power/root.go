package power

import (
	"github.com/spf13/cobra"
	nodepower "github.com/warewulf/warewulf/internal/app/wwctl/node/power"
)

const deprecation = "use 'wwctl node power ACTION' instead"

// Deprecated: use 'wwctl node power' commands instead.
// This is kept purely for backward compatibility.
func GetCommand() *cobra.Command {
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "power COMMAND [OPTIONS]",
		Short:                 "Warewulf node power management",
		Long:                  "This command controls the power state of nodes.",
		Args:                  cobra.NoArgs,
		Deprecated:            deprecation,
	}
	for _, action := range nodepower.Actions {
		actionCmd := nodepower.GetActionCommand(action.Name)
		actionCmd.Deprecated = deprecation
		baseCmd.AddCommand(actionCmd)
	}
	return baseCmd
}
