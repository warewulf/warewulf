package configure

import (
	"fmt"
	"os"

	"github.com/hpcng/warewulf/internal/app/wwctl/configure/dhcp"
	"github.com/hpcng/warewulf/internal/app/wwctl/configure/hosts"
	"github.com/hpcng/warewulf/internal/app/wwctl/configure/nfs"
	"github.com/hpcng/warewulf/internal/app/wwctl/configure/ssh"
	"github.com/hpcng/warewulf/internal/app/wwctl/configure/tftp"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:   "configure",
		Short: "Manage system services",
		Long: "This application allows you to manage and initialize Warewulf dependent system\n" +
			"services based on the configuration in the warewulf.conf file.",
		RunE: CobraRunE,
	}
	SetDoAll bool
)

func init() {
	baseCmd.AddCommand(dhcp.GetCommand())
	baseCmd.AddCommand(tftp.GetCommand())
	baseCmd.AddCommand(hosts.GetCommand())
	baseCmd.AddCommand(ssh.GetCommand())
	baseCmd.AddCommand(nfs.GetCommand())

	baseCmd.PersistentFlags().BoolVarP(&SetDoAll, "all", "a", false, "Configure all services")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}

func CobraRunE(cmd *cobra.Command, args []string) error {

	if SetDoAll {
		fmt.Printf("################################################################################\n")
		fmt.Printf("Configuring: DHCP\n")
		dhcp.Configure(false)

		fmt.Printf("################################################################################\n")
		fmt.Printf("Configuring: TFTP\n")
		tftp.Configure(false)

		fmt.Printf("################################################################################\n")
		fmt.Printf("Configuring: /etc/hosts\n")
		hosts.Configure(false)

		fmt.Printf("################################################################################\n")
		fmt.Printf("Configuring: NFS\n")
		nfs.Configure(false)

		fmt.Printf("################################################################################\n")
		fmt.Printf("Configuring: SSH\n")
		ssh.Configure(false)

	} else {
		cmd.Help()
		os.Exit(0)
	}
	return nil
}
