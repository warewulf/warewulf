package dhcp

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"net"
	"os"
	"text/template"
)

type dhcpTemplate struct {
	Ipaddr     string
	RangeStart string
	RangeEnd   string
	Netmask    string
	Nodes      []node.NodeInfo
}

func CobraRunE(cmd *cobra.Command, args []string) error {

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	controllers, err := nodeDB.FindAllControllers()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not find all controllers: %s\n", err)
		os.Exit(1)
	}

	for _, controller := range controllers {
		var d dhcpTemplate
		var configured bool

		addrs, err := net.InterfaceAddrs()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not get network interfaces: %s\n", err)
			os.Exit(1)
		}

		for _, address := range addrs {
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.String() == controller.Ipaddr {
					mask := ipnet.IP.DefaultMask()
					d.Ipaddr = ipnet.IP.String()
					d.Netmask = fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3])
					d.RangeStart = controller.Services.Dhcp.RangeStart
					d.RangeEnd = controller.Services.Dhcp.RangeEnd
					configured = true
					fmt.Printf("%#v\n", d)
					break
				}
			}
		}

		if configured == false {
			wwlog.Printf(wwlog.ERROR, "Could not identify this system in the Warewulf configuration by it's IP address\n")
			os.Exit(1)
		}

		if controller.Services.Dhcp.ConfigFile == "" {
			wwlog.Printf(wwlog.ERROR, "Could not locate the DHCP configuration file for this controller\n")
			os.Exit(1)
		}

		if _, ok := nodeDB.Controllers[controller.Id]; !ok {
			wwlog.Printf(wwlog.ERROR, "We should never get here, but since we did, Hello! %s\n", err)
			os.Exit(1)
		}

		nodes, err := nodeDB.FindAllNodes()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not find all controllers: %s\n", err)
			os.Exit(1)
		}

		for _, node := range nodes {
			d.Nodes = append(d.Nodes, node)
		}

		tmpl, err := template.New("default-dhcpd.conf").ParseFiles("/etc/warewulf/dhcp/default-dhcpd.conf")
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}

		//		w, err := os.OpenFile(controller.Services.Dhcp.ConfigFile, os.O_RDWR|os.O_CREATE, 0640)
		//		if err != nil {
		//			wwlog.Printf(wwlog.ERROR, "%s\n", err)
		//			os.Exit(1)
		//		}
		//		defer w.Close()

		err = tmpl.Execute(os.Stdout, d)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}

		// Just in case we get here, we've now finished the loop
		break
	}

	return nil
}
