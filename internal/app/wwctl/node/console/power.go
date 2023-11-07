package console

import (
	"os"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/power"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/hpcng/warewulf/pkg/hostlist"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var returnErr error = nil

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Error("Could not open node configuration: %s", err)
		os.Exit(1)
	}

	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		wwlog.Error("Could not get node list: %s", err)
		os.Exit(1)
	}

	args = hostlist.Expand(args)

	if len(args) > 0 {
		nodes = node.FilterByName(nodes, args)
	} else {
		//nolint:errcheck
		cmd.Usage()
		os.Exit(1)
	}

	if len(nodes) == 0 {
		wwlog.Info("No nodes found\n")
		os.Exit(1)
	}

	for _, node := range nodes {

		if node.Ipmi.Ipaddr.Get() == "" {
			wwlog.Error("%s: No IPMI IP address", node.Id.Get())
			continue
		}

		ipmiCmd := power.IPMI{
			NodeName:  node.Id.Get(),
			HostName:  node.Ipmi.Ipaddr.Get(),
			Port:      node.Ipmi.Port.Get(),
			User:      node.Ipmi.UserName.Get(),
			Password:  node.Ipmi.Password.Get(),
			AuthType:  "MD5",
			Interface: node.Ipmi.Interface.Get(),
		}

		err := ipmiCmd.Console()

		if err != nil {
			wwlog.Error("%s: Console problem", node.Id.Get())
			returnErr = err
			continue
		}

	}

	return returnErr
}
