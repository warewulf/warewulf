package nfs

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "nfs [OPTIONS]",
		Short:                 "Manage and initialize NFS",
		Long: "NFS is an optional dependent service of Warewulf, this tool will automatically\n" +
			"configure NFS as per the configuration in the warewulf.conf file.",
		RunE: CobraRunE,
	}
	setShow bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&setShow, "show", "s", false, "Show configuration (don't update)")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
