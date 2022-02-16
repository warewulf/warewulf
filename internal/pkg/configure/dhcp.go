package configure

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"

	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
)

func configureDHCP(show bool) error {

	var (
		d struct {
			Ipaddr     string
			Port       string
			RangeStart string
			RangeEnd   string
			Network    string
			Netmask    string
			Nodes      []node.NodeInfo
		}

		templateFile string
	)

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

	if !controller.Dhcp.Enabled {
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

	d.Nodes = append(d.Nodes, nodes...)

	templateFile = dhcpTemplateFile(controller.Dhcp.Template)
	tmpl, err := template.New(path.Base(templateFile)).ParseFiles(templateFile)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	d.Ipaddr = controller.Ipaddr
	d.Port = strconv.Itoa(controller.Warewulf.Port)
	d.Network = controller.Network
	d.Netmask = controller.Netmask
	d.RangeStart = controller.Dhcp.RangeStart
	d.RangeEnd = controller.Dhcp.RangeEnd

	if !show {
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
		err = util.SystemdStart(controller.Dhcp.SystemdName)
		if err != nil {
			return errors.Wrap(err, "failed to start")
		}
	} else {
		err = tmpl.Execute(os.Stdout, d)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}

	}

	return nil
}

// dhcpTemplateFile returns the path of the warewulf dhcp template given controller.Dhcp.Template.
func dhcpTemplateFile(controllerDhcpTemplate string) (templateFile string) {
	if controllerDhcpTemplate == "" {
		templateFile = path.Join(buildconfig.SYSCONFDIR(), "warewulf/dhcp/default-dhcpd.conf")
	} else {
		if strings.HasPrefix(controllerDhcpTemplate, "/") {
			templateFile = controllerDhcpTemplate
		} else {
			templateFile = path.Join(buildconfig.SYSCONFDIR(), "warewulf/dhcp/"+controllerDhcpTemplate+"-dhcpd.conf")
		}
	}
	return
}
