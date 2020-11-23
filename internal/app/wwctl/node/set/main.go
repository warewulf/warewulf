package set

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"os"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var err error
	var nodes []node.NodeInfo

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Cloud not get nodeList: %s\n", err)
		os.Exit(1)
	}

	if len(args) > 0 {
		nodes, err = nodeDB.SearchByNameList(args)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Cloud not get nodeList: %s\n", err)
			os.Exit(1)
		}

	} else {
		cmd.Usage()
		os.Exit(1)
	}

	for _, n := range nodes {
		wwlog.Printf(wwlog.VERBOSE, "Evaluating node: %s\n", n.Fqdn)
		if SetVnfs != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting vnfs to: %s\n", n.Fqdn, SetVnfs)
			err := nodeDB.SetNodeVal(n.Gid, n.Id, "vnfs", SetVnfs)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetKernel != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting kernel to: %s\n", n.Fqdn, SetVnfs)
			err := nodeDB.SetNodeVal(n.Gid, n.Id, "kernel", SetKernel)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetDomainName != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting domain name to: %s\n", n.Fqdn, SetDomainName)
			err := nodeDB.SetNodeVal(n.Gid, n.Id, "domain", SetDomainName)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetIpxe != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting iPXE template to: %s\n", n.Fqdn, SetIpxe)
			err := nodeDB.SetNodeVal(n.Gid, n.Id, "ipxe", SetIpxe)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetRuntimeOverlay != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting runtime overlay to: %s\n", n.Fqdn, SetRuntimeOverlay)
			err := nodeDB.SetNodeVal(n.Gid, n.Id, "runtimeoverlay", SetRuntimeOverlay)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetSystemOverlay != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting system overlay to: %s\n", n.Fqdn, SetSystemOverlay)
			err := nodeDB.SetNodeVal(n.Gid, n.Id, "systemoverlay", SetSystemOverlay)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetHostname != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting hostname to: %s\n", n.Fqdn, SetHostname)
			err := nodeDB.SetNodeVal(n.Gid, n.Id, "hostname", SetHostname)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetIpmiIpaddr != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting IPMI IP address to: %s\n", n.Fqdn, SetIpmiIpaddr)
			err := nodeDB.SetNodeVal(n.Gid, n.Id, "ipmiipaddr", SetIpmiIpaddr)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetIpmiUsername != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting IPMI IP username to: %s\n", n.Fqdn, SetIpmiUsername)
			err := nodeDB.SetNodeVal(n.Gid, n.Id, "ipmiusername", SetIpmiUsername)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetIpmiPassword != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting IPMI IP password to: %s\n", n.Fqdn, SetIpmiPassword)
			err := nodeDB.SetNodeVal(n.Gid, n.Id, "ipmipassword", SetIpmiPassword)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}



		if SetNetDevDel == true {
			if SetNetDev == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netdev' option\n")
				os.Exit(1)
			}
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Deleting network device: %s\n", n.Fqdn, SetNetDev)
			err := nodeDB.DelNodeNet(n.Gid, n.Id, SetNetDev)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}

		}
		if SetIpaddr != "" {
			if SetNetDev == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netdev' option\n")
				os.Exit(1)
			}
			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting Ipaddr to: %s\n", n.Fqdn, SetNetDev, SetIpaddr)
			err := nodeDB.SetNodeNet(n.Gid, n.Id, SetNetDev, "ipaddr", SetIpaddr)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetNetmask != "" {
			if SetNetDev == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netdev' option\n")
				os.Exit(1)
			}
			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting netmask to: %s\n", n.Fqdn, SetNetDev, SetNetmask)
			err := nodeDB.SetNodeNet(n.Gid, n.Id, SetNetDev, "netmask", SetNetmask)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetGateway != "" {
			if SetNetDev == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netdev' option\n")
				os.Exit(1)
			}
			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting gateway to: %s\n", n.Fqdn, SetNetDev, SetGateway)
			err := nodeDB.SetNodeNet(n.Gid, n.Id, SetNetDev, "gateway", SetGateway)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetHwaddr != "" {
			if SetNetDev == "" {
				wwlog.Printf(wwlog.ERROR, "You must include the '--netdev' option\n")
				os.Exit(1)
			}
			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting HW address to: %s\n", n.Fqdn, SetNetDev, SetHwaddr)
			err := nodeDB.SetNodeNet(n.Gid, n.Id, SetNetDev, "hwaddr", SetHwaddr)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
	}

	if len(nodes) > 0 {
		q := fmt.Sprintf("Are you sure you want to modify %d nodes(s)", len(nodes))

		prompt := promptui.Prompt{
			Label:     q,
			IsConfirm: true,
		}

		result, _ := prompt.Run()

		if result == "y" || result == "yes" {
			nodeDB.Persist()
		}

	} else {
		fmt.Printf("No nodes found\n")
	}

	return nil
}