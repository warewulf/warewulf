package list

import (
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/spf13/cobra"
)

type variables struct {
	output string
}

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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return util.ValidOutput(vars.output)
		},
	}

	baseCmd.PersistentFlags().StringVarP(&vars.output, "output", "o", "text", "output format `json | text | yaml | csv`")
	return baseCmd
}
