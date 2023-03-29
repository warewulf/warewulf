package list

import "github.com/spf13/cobra"

type variables struct{}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	vars := variables{}
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "list [OPTIONS]",
		Short:                 "List imported Warewulf containers",
		Long:                  "This command will show you the containers that are imported into Warewulf.",
		RunE:                  CobraRunE(&vars),
		Aliases:               []string{"ls"},
	}
	return baseCmd
}
