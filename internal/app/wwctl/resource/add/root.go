package add

import (
	"github.com/spf13/cobra"
)

type variables struct {
	tags     map[string]string
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	vars := variables{}
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "add PROFILE",
		Short:                 "Add a new node profile",
		Long:                  "This command adds a new named PROFILE.",
		Aliases:               []string{"new", "create"},
		RunE:                  CobraRunE(&vars),
		Args:                  cobra.ExactArgs(1),
	}
	baseCmd.PersistentFlags().StringToStringVar(&vars.tags, "restagadd", map[string]string{}, "add cluster wide resource tags")
	return baseCmd
}
