package dhcp

import (
	"fmt"
	"github.com/brotherpowers/ipsubnet"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"net"
	"os"
	"path"
	"strings"
	"text/template"
)

type dhcpTemplate struct {
	Ipaddr     string
	RangeStart string
	RangeEnd   string
	Network    string
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
		var templateFile string
		var configWriter *os.File
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
					m, _ := ipnet.Mask.Size()
					sub := ipsubnet.SubnetCalculator(ipnet.IP.String(), m)

					d.Ipaddr = ipnet.IP.String()
					d.Network = sub.GetNetworkPortion()
					d.Netmask = sub.GetSubnetMask()
					d.RangeStart = controller.Services.Dhcp.RangeStart
					d.RangeEnd = controller.Services.Dhcp.RangeEnd
					configured = true
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

		if controller.Services.Dhcp.Template == "" {
			templateFile = "/etc/warewulf/dhcp/default-dhcpd.conf"
		} else {
			if strings.HasPrefix(controller.Services.Dhcp.Template, "/") {
				templateFile = controller.Services.Dhcp.Template
			} else {
				templateFile = fmt.Sprintf("/etc/warewulf/dhcp/%s-dhcpd.conf", controller.Services.Dhcp.Template)
			}
		}

		tmpl, err := template.New(path.Base(templateFile)).ParseFiles(templateFile)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}

		if ShowConfig == true {
			configWriter = os.Stdout
		} else {
			configWriter, err = os.OpenFile(controller.Services.Dhcp.ConfigFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0640)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
			defer configWriter.Close()
		}

		err = tmpl.Execute(configWriter, d)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}

		util.ExecInteractive("/bin/sh", "-c", controller.Services.Dhcp.EnableCmd)
		util.ExecInteractive("/bin/sh", "-c", controller.Services.Dhcp.RestartCmd)

		// Just in case we get here, we've now finished the loop
		break
	}

	return nil
}
