package service

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/app/wwctl/service/dhcp"
	"github.com/hpcng/warewulf/internal/app/wwctl/service/tftp"
	"github.com/spf13/cobra"
	"os"
)

var (
	baseCmd = &cobra.Command{
		Use:   "service",
		Short: "Initialize Warewulf services",
		Long:  "Warewulf Service Initialization",
		RunE:  CobraRunE,
	}
	SetDoAll bool
)

func init() {
	baseCmd.AddCommand(dhcp.GetCommand())
	baseCmd.AddCommand(tftp.GetCommand())
	baseCmd.PersistentFlags().BoolVarP(&SetDoAll, "all", "a", false, "Configure all services")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}

func CobraRunE(cmd *cobra.Command, args []string) error {

	if SetDoAll == true {
		fmt.Printf("################################################################################\n")
		dhcp.Configure(false)

		fmt.Printf("################################################################################\n")
		tftp.Configure(false)
	} else {
		cmd.Help()
		os.Exit(0)
	}
	return nil
}
