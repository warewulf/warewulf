package poweroff

import (
	"github.com/hpcng/warewulf/internal/pkg/assets"
	"github.com/hpcng/warewulf/internal/pkg/power"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	var returnErr error = nil

	var nodeList []assets.NodeInfo

	if len(args) >= 1 {
		//nodeList, _ = assets.SearchByName(args[0])
		nodeList, _ = assets.SearchByNameList(args)
	} else {
		wwlog.Printf(wwlog.ERROR, "No requested nodes\n")
		os.Exit(255)
	}

	if len(nodeList) == 0 {
		wwlog.Printf(wwlog.ERROR, "No nodes found matching: '%s'\n", args[0])
		os.Exit(255)
	} else {
		wwlog.Printf(wwlog.VERBOSE, "Found %d matching nodes for power command\n", len(nodeList))
	}

	for _, node := range nodeList {

		if node.IpmiIpaddr == "" {
			wwlog.Printf(wwlog.ERROR, "%s: No IPMI IP address\n", node.HostName)
			continue
		}

		ipmiCmd := power.IPMI{
			HostName: node.IpmiIpaddr,
			User:     "ADMIN",
			Password: "ADMIN",
			AuthType: "MD5",
		}

		out, err := ipmiCmd.PowerOff()

		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s: %s\n", node.HostName, out)
			returnErr = err
			continue
		}

		wwlog.Printf(wwlog.INFO, "%s: %s\n", node.HostName, out)
	}

	return returnErr
}
