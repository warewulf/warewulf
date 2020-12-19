package hosts

import (
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
	"text/template"
)

type TemplateStruct struct {
	Ipaddr   string
	Fqdn     string
	AllNodes []node.NodeInfo
}

func CobraRunE(cmd *cobra.Command, args []string) error {
	var replace TemplateStruct

	if util.IsFile("/etc/warewulf/hosts.tmpl") == false {
		wwlog.Printf(wwlog.WARN, "Template not found, not updating host file\n")
		return nil
	}

	controller, err := warewulfconf.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	n, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	tmpl, err := template.ParseFiles("/etc/warewulf/hosts.tmpl")
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not parse hosts template: %s\n", err)
		os.Exit(1)
	}

	w, err := os.OpenFile("/etc/hosts", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}
	defer w.Close()

	nodes, _ := n.FindAllNodes()

	replace.AllNodes = nodes
	replace.Ipaddr = controller.Ipaddr
	replace.Fqdn = controller.Fqdn

	if SetShow == false {
		err = tmpl.Execute(w, replace)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}
	} else {
		err = tmpl.Execute(os.Stdout, replace)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}

	}

	return nil
}
