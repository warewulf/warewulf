package dhcp

import (
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
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
	if SetShow == false && SetPersist == false {
		fmt.Println(cmd.Help())
		os.Exit(0)
	}

	return Configure(SetShow)
}

func Configure(show bool) error {
	var d dhcpTemplate
	var templateFile string

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	controller, err := warewulfconf.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	if controller.Ipaddr == "" {
		wwlog.Printf(wwlog.ERROR, "The Warewulf IP Address is not properly configured\n")
		os.Exit(1)
	}

	if controller.Netmask == "" {
		wwlog.Printf(wwlog.ERROR, "The Warewulf Netmask is not properly configured\n")
		os.Exit(1)
	}

	if controller.Dhcp.Enabled == false {
		wwlog.Printf(wwlog.INFO, "This system is not configured as a Warewulf DHCP controller\n")
		os.Exit(1)
	}

	if controller.Dhcp.RangeStart == "" {
		wwlog.Printf(wwlog.ERROR, "Configuration is not defined: `dhcpd range start`\n")
		os.Exit(1)
	}

	if controller.Dhcp.RangeEnd == "" {
		wwlog.Printf(wwlog.ERROR, "Configuration is not defined: `dhcpd range end`\n")
		os.Exit(1)
	}

	if controller.Dhcp.ConfigFile == "" {
		controller.Dhcp.ConfigFile = "/etc/dhcp/dhcpd.conf"
	}

	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not find all controllers: %s\n", err)
		os.Exit(1)
	}

	for _, node := range nodes {
		d.Nodes = append(d.Nodes, node)
	}

	if controller.Dhcp.Template == "" {
		templateFile = "/etc/warewulf/dhcp/default-dhcpd.conf"
	} else {
		if strings.HasPrefix(controller.Dhcp.Template, "/") {
			templateFile = controller.Dhcp.Template
		} else {
			templateFile = fmt.Sprintf("/etc/warewulf/dhcp/%s-dhcpd.conf", controller.Dhcp.Template)
		}
	}

	tmpl, err := template.New(path.Base(templateFile)).ParseFiles(templateFile)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	d.Ipaddr = controller.Ipaddr
	d.Network = controller.Network
	d.Netmask = controller.Netmask
	d.RangeStart = controller.Dhcp.RangeStart
	d.RangeEnd = controller.Dhcp.RangeEnd

	if show == false {
		fmt.Printf("Writing the DHCP configuration file\n")
		configWriter, err := os.OpenFile(controller.Dhcp.ConfigFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0640)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}
		defer configWriter.Close()
		err = tmpl.Execute(configWriter, d)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}

		fmt.Printf("Enabling and restarting the DHCP services\n")
		util.SystemdStart(controller.Dhcp.SystemdName)

	} else {
		err = tmpl.Execute(os.Stdout, d)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}

	}

	return nil
}
