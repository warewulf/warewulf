package hosts

import (
	"bytes"
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
	"text/template"
)

type TemplateStruct struct {
	PrevHostFile string
	Ipaddr       string
	Fqdn         string
	AllNodes     []node.NodeInfo
}

func CobraRunE(cmd *cobra.Command, args []string) error {
	var replace TemplateStruct

	if SetShow == false && SetPersist == false {
		fmt.Println(cmd.Help())
		os.Exit(0)
	}

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

	replace.PrevHostFile = ""
	w, err := os.Open("/etc/hosts")
	if err != nil {
		wwlog.Printf(wwlog.WARN, "%s\n", err)
	} else {
		// if /etc/hosts.ww does not exist, backup /etc/hosts to /etc/hosts.wwbackup
		if !util.IsFile("/etc/hosts.wwbackup") {
			err = util.CopyFile("/etc/hosts", "/etc/hosts.wwbackup")
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
			}
		}

		// read all lines before the # warewulf comment and put into PrevHostFile template variable
		lines, _ := util.ReadFile("/etc/hosts")
		if lines != nil {
			var buffer bytes.Buffer
			for _, line := range lines {
				//wwlog.Printf(wwlog.INFO, "Reading line: %s\n", line)
				if util.ValidString(line, "^#.*maintained by warewulf") {
					break
				}
				buffer.WriteString(line)
				buffer.WriteString("\n")
			}
			replace.PrevHostFile = buffer.String()
		}
	}

	//wwlog.Printf(wwlog.INFO, "PrevHostFile is %s\n", replace.PrevHostFile)

	w.Close()

	nodes, _ := n.FindAllNodes()

	replace.AllNodes = nodes
	replace.Ipaddr = controller.Ipaddr
	replace.Fqdn = controller.Fqdn

	if SetShow == false {
		// only open "/etc/hosts" when intended to write, as 'os.O_TRUNC' will empty the file otherwise.
		w, err = os.OpenFile("/etc/hosts", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}
		defer w.Close()

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
