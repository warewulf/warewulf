package cacheclean

import "github.com/spf13/cobra"

type variables struct {
	garbageCollect bool
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	vars := variables{}
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "clean [OPTIONS]",
		Short:                 "Clean blobs in oci cache",
		Long:                  "This command will clean up the blobs in the oci cache",
		RunE:                  CobraRunE(&vars),
	}
	baseCmd.Flags().BoolVarP(&vars.garbageCollect, "gc", "g", false, "run garbage collector")
	return baseCmd
}
