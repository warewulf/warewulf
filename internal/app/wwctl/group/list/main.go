package list

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	groups, err := nodeDB.FindAllGroups()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not find all nodes: %s\n", err)
		os.Exit(1)
	}

	if ShowAll == true {
		for _, group := range groups {
			fmt.Printf("################################################################################\n")
			fmt.Printf("%-20s %-18s %8s: %s\n", group.Id.Get(), "Id", group.Id.Source(), group.Id.Print())
			fmt.Printf("%-20s %-18s %8s: %s\n", group.Id.Get(), "Controller", group.Cid.Source(), group.Cid.Print())
			fmt.Printf("%-20s %-18s %8s: %s\n", group.Id.Get(), "DomainName", group.DomainName.Source(), group.DomainName.Print())
			fmt.Printf("%-20s %-18s %8s: %s\n", group.Id.Get(), "VNFS", group.Vnfs.Source(), group.Vnfs.Print())
			fmt.Printf("%-20s %-18s %8s: %s\n", group.Id.Get(), "KernelVersion", group.KernelVersion.Source(), group.KernelVersion.Print())
			fmt.Printf("%-20s %-18s %8s: %s\n", group.Id.Get(), "KernelArgs", group.KernelArgs.Source(), group.KernelArgs.Print())
			fmt.Printf("%-20s %-18s %8s: %s\n", group.Id.Get(), "RuntimeOverlay", group.RuntimeOverlay.Source(), group.RuntimeOverlay.Print())
			fmt.Printf("%-20s %-18s %8s: %s\n", group.Id.Get(), "SystemOverlay", group.SystemOverlay.Source(), group.SystemOverlay.Print())
			fmt.Printf("%-20s %-18s %8s: %s\n", group.Id.Get(), "IPMI Netmask", group.IpmiNetmask.Source(), group.IpmiNetmask.Print())
			fmt.Printf("%-20s %-18s %8s: %s\n", group.Id.Get(), "IPMI UserName", group.IpmiUserName.Source(), group.IpmiUserName.Print())
			fmt.Printf("%-20s %-18s %8s: %s\n", group.Id.Get(), "IPMI Password", group.IpmiPassword.Source(), group.IpmiPassword.Print())
			fmt.Printf("%-20s %-18s %8s: %s\n", group.Id.Get(), "Ipxe", group.Ipxe.Source(), group.Ipxe.Print())
			fmt.Printf("%-20s %-18s %8s: %s\n", group.Id.Get(), "Profiles", "group", strings.Join(group.Profiles, ","))

		}
	} else {
		fmt.Printf("%-22s %-16s %-16s %s\n", "GROUP NAME", "DOMAINNAME", "CONTROLLER", "PROFILES")
		for _, g := range groups {
			fmt.Printf("%-22s %-16s %-16s %s\n", g.Id.Get(), g.DomainName.Get(), g.Cid.Get(), strings.Join(g.Profiles, ","))
		}
	}

	return nil
}
