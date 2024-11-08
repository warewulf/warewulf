package console

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/power"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var returnErr error = nil

	nodeDB, err := node.New()
	if err != nil {
		return fmt.Errorf("could not open node configuration: %s", err)
	}

	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		return fmt.Errorf("could not get node list: %s", err)
	}

	args = hostlist.Expand(args)

	if len(args) > 0 {
		nodes = node.FilterNodeListByName(nodes, args)
	} else {
		//nolint:errcheck
		cmd.Usage()
		os.Exit(1)
	}

	if len(nodes) == 0 {
		return fmt.Errorf("no nodes found")
	}

	for _, node := range nodes {

		if node.Ipmi.Ipaddr.IsUnspecified() {
			wwlog.Error("%s: No IPMI IP address", node.Id())
			continue
		}
		ipmiCmd := power.IPMI{IpmiConf: *node.Ipmi}
		err := ipmiCmd.Console()
		if err != nil {
			wwlog.Error("%s: Console problem", node.Id())
			returnErr = err
			continue
		}

	}

	return returnErr
}
