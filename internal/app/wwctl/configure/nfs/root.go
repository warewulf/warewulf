package nfs

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "nfs [OPTIONS]",
		Short:                 "Manage and initialize NFS",
		Long: "NFS is an optional dependent service of Warewulf, this tool will automatically\n" +
			"configure NFS as per the configuration in the warewulf.conf file.",
		RunE:              CobraRunE,
		Args:              cobra.NoArgs,
		ValidArgsFunction: completions.None,
	}
)

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
