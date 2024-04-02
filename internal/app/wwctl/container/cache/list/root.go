package cachelist

import "github.com/spf13/cobra"

type variables struct {
	allblobs     bool
	showManifest bool
	showDigest   bool
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	vars := variables{}
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "list [OPTIONS]",
		Short:                 "List blobs in oci cache",
		Long:                  "This command will show you the contents of the oci cache",
		RunE:                  CobraRunE(&vars),
		Aliases:               []string{"ls"},
	}
	baseCmd.Flags().BoolVarP(&vars.allblobs, "allblobs", "a", false, "list all blobs")
	baseCmd.Flags().BoolVarP(&vars.showManifest, "fullmanifest", "m", false, "show also manifests")
	baseCmd.Flags().BoolVarP(&vars.showDigest, "digest", "f", false, "show also digest")
	return baseCmd
}
