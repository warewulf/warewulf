package service

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/service/dhcp"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:   "service",
		Short: "Initialize Warewulf services",
		Long:  "Warewulf Service Initialization",
	}
)

func init() {
	baseCmd.AddCommand(dhcp.GetCommand())
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
