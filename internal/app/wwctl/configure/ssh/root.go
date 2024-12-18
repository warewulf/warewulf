package ssh

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "ssh [OPTIONS]",
		Short:                 "Manage and initialize SSH",
		Long: "SSH is an optionally dependent service for Warewulf, this tool will automatically\n" +
			"setup the ssh keys nodes using the 'default' system overlay as well as user owned\n" +
			"keys.",
		RunE: CobraRunE,
	}
	keyTypes []string
)

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	baseCmd.PersistentFlags().StringArrayVarP(&keyTypes, "keytypes", "t", []string{}, "ssh key types to be created")
	return baseCmd
}
