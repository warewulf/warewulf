package list

import "C"
import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	controllers, err := nodeDB.FindAllControllers()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not find all nodes: %s\n", err)
		os.Exit(1)
	}

	if ShowAll == true {
		for _, c := range controllers {
			fmt.Printf("################################################################################\n")
			fmt.Printf("%-15s %15s : %s = %s\n", c.Id, "", "IP Address", c.Ipaddr)
			fmt.Printf("%-15s %15s : %s = %s\n", c.Id, "", "Domain Name", c.DomainName)
			fmt.Printf("%-15s %15s : %s = %s\n", c.Id, "", "FQDN", c.Fqdn)
			fmt.Printf("%-15s %15s : %s = %s\n", c.Id, "", "Comment", c.Comment)
			fmt.Printf("%-15s %15s : %s = %s\n", c.Id, "Warewulf", "Port", c.Services.Warewulfd.Port)
			fmt.Printf("%-15s %15s : %s = %t\n", c.Id, "Warewulf", "Secure", c.Services.Warewulfd.Secure)
			fmt.Printf("%-15s %15s : %s = %s\n", c.Id, "Warewulf", "Enable CMD", c.Services.Warewulfd.EnableCmd)
			fmt.Printf("%-15s %15s : %s = %s\n", c.Id, "Warewulf", "Restart CMD", c.Services.Warewulfd.RestartCmd)
			fmt.Printf("%-15s %15s : %s = %t\n", c.Id, "DHCPD", "Enabled", c.Services.Dhcp.Enabled)
			if c.Services.Dhcp.Enabled == true {
				fmt.Printf("%-15s %15s : %s = %s\n", c.Id, "DHCPD", "Template", c.Services.Dhcp.Template)
				fmt.Printf("%-15s %15s : %s = %s\n", c.Id, "DHCPD", "ConfigFile", c.Services.Dhcp.ConfigFile)
				fmt.Printf("%-15s %15s : %s = %s\n", c.Id, "DHCPD", "RangeStart", c.Services.Dhcp.RangeStart)
				fmt.Printf("%-15s %15s : %s = %s\n", c.Id, "DHCPD", "RangeEnd", c.Services.Dhcp.RangeEnd)
				fmt.Printf("%-15s %15s : %s = %s\n", c.Id, "DHCPD", "Enable CMD", c.Services.Dhcp.EnableCmd)
				fmt.Printf("%-15s %15s : %s = %s\n", c.Id, "DHCPD", "Restart CMD", c.Services.Dhcp.RestartCmd)
			}
			fmt.Printf("%-15s %15s : %s = %t\n", c.Id, "TFTP", "Enabled", c.Services.Tftp.Enabled)
			if c.Services.Tftp.Enabled == true {
				fmt.Printf("%-15s %15s : %s = %s\n", c.Id, "TFTP", "TftpRoot", c.Services.Tftp.TftpRoot)
				fmt.Printf("%-15s %15s : %s = %s\n", c.Id, "TFTP", "Enable CMD", c.Services.Tftp.EnableCmd)
				fmt.Printf("%-15s %15s : %s = %s\n", c.Id, "TFTP", "Restart CMD", c.Services.Tftp.RestartCmd)
			}
			fmt.Printf("%-15s %15s : %s = %t\n", c.Id, "NFS", "Enabled", c.Services.Nfs.Enabled)
			if c.Services.Nfs.Enabled == true {
				for _, e := range c.Services.Nfs.Exports {
					fmt.Printf("%-15s %15s : %s = %s\n", c.Id, "NFS", "Exports", e)
				}
				fmt.Printf("%-15s %15s : %s = %s\n", c.Id, "NFS", "Enable CMD", c.Services.Nfs.EnableCmd)
				fmt.Printf("%-15s %15s : %s = %s\n", c.Id, "NFS", "Restart CMD", c.Services.Nfs.RestartCmd)
			}

		}
	} else {
		fmt.Printf("%-22s\n", "CONTROLLER NAME")
		for _, c := range controllers {
			fmt.Printf("%-22s\n", c.Id)
		}
	}

	return nil
}
