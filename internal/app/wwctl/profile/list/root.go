package list

import "github.com/spf13/cobra"

type variables struct {
	showAll     bool
	showFullAll bool
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	vars := variables{}
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "list [OPTIONS] [PROFILE ...]",
		Short:                 "List profiles and configurations",
		Long:                  "This command will display configurations for PROFILE.",
		RunE:                  CobraRunE(&vars),
		Aliases:               []string{"ls"},
	}
	baseCmd.PersistentFlags().BoolVarP(&vars.showAll, "all", "a", false, "Show all profile configurations")
	baseCmd.PersistentFlags().BoolVarP(&vars.showFullAll, "fullall", "A", false, "Show all profile configurations inclusive empty entries")

	return baseCmd
}
