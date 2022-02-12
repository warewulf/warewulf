package configure

import (
	"os"

	"github.com/hpcng/warewulf/internal/app/wwctl/configure/dhcp"
	"github.com/hpcng/warewulf/internal/app/wwctl/configure/hosts"
	"github.com/hpcng/warewulf/internal/app/wwctl/configure/nfs"
	"github.com/hpcng/warewulf/internal/app/wwctl/configure/ssh"
	"github.com/hpcng/warewulf/internal/app/wwctl/configure/tftp"
	"github.com/hpcng/warewulf/internal/pkg/configure"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "configure [OPTIONS]",
		Short:                 "Manage system services",
		Long: "This application allows you to manage and initialize Warewulf dependent system\n" +
			"services based on the configuration in the warewulf.conf file.",
		RunE: CobraRunE,
	}
	allFunctions bool
)

func init() {
	baseCmd.AddCommand(dhcp.GetCommand())
	baseCmd.AddCommand(tftp.GetCommand())
	baseCmd.AddCommand(hosts.GetCommand())
	baseCmd.AddCommand(ssh.GetCommand())
	baseCmd.AddCommand(nfs.GetCommand())

	baseCmd.PersistentFlags().BoolVarP(&allFunctions, "all", "a", false, "Configure all services")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}

func CobraRunE(cmd *cobra.Command, args []string) error {
	var err error
	if allFunctions {
		for _, s := range [5]string{"DHPC", "hosts", "NFS", "SSH", "TFTP"} {
			err = configure.Configure(s, false)
			if err != nil {
				os.Exit(1)
			}

		}
	} else {
		_ = cmd.Help()
		os.Exit(0)
	}

	return nil
}
