package console

import (
	"os"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/power"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var returnErr error = nil
	var nodeList []node.NodeInfo

	n, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	if len(args) == 1 {
		nodeList, _ = n.SearchByName(args[0])
	} else {
		wwlog.Printf(wwlog.ERROR, "Please specify a node\n")
		os.Exit(255)
	}

	if len(nodeList) == 0 {
		wwlog.Printf(wwlog.ERROR, "No nodes found matching: '%s'\n", args[0])
		os.Exit(255)
	}

	for _, node := range nodeList {

		if node.IpmiIpaddr.Get() == "" {
			wwlog.Printf(wwlog.ERROR, "%s: No IPMI IP address\n", node.Id.Get())
			continue
		}

		ipmiCmd := power.IPMI{
			NodeName: node.Id.Get(),
			HostName: node.IpmiIpaddr.Get(),
			User:     node.IpmiUserName.Get(),
			Password: node.IpmiPassword.Get(),
			AuthType: "MD5",
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
