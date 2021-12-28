package console

import (
	"fmt"
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
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not get node list: %s\n", err)
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
		fmt.Printf("No nodes found\n")
		os.Exit(1)
	}

	for _, node := range nodes {

		if node.IpmiIpaddr.Get() == "" {
			wwlog.Printf(wwlog.ERROR, "%s: No IPMI IP address\n", node.Id.Get())
			continue
		}

		ipmiCmd := power.IPMI{
			NodeName:  node.Id.Get(),
			HostName:  node.IpmiIpaddr.Get(),
			Port:      node.IpmiPort.Get(),
			User:      node.IpmiUserName.Get(),
			Password:  node.IpmiPassword.Get(),
			AuthType:  "MD5",
			Interface: node.IpmiInterface.Get(),
		}

		err := ipmiCmd.Console()

		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s: Console problem\n", node.Id.Get())
			returnErr = err
			continue
		}

	}

	return returnErr
}
