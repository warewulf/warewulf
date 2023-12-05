package powerstatus

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/batch"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/power"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
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

	if len(args) > 0 {
		nodes = node.FilterByName(nodes, hostlist.Expand(args))
	} else {
		//nolint:errcheck
		cmd.Usage()
		os.Exit(1)
	}

	if len(nodes) == 0 {
		fmt.Printf("No nodes found\n")
		os.Exit(1)
	}

	batchpool := batch.New(50)
	jobcount := len(nodes)
	results := make(chan power.IPMI, jobcount)

	for _, n := range nodes {

		if n.Ipmi.Ipaddr.Get() == "" {
			wwlog.Error("%s: No IPMI IP address", n.Id.Get())
			continue
		}
		var conf node.NodeConf
		conf.GetFrom(n)
		ipmiCmd := power.IPMI{IpmiConf: *conf.Ipmi}
		batchpool.Submit(func() {
			//nolint:errcheck
			ipmiCmd.PowerStatus()
			results <- ipmiCmd
		})

	}

	batchpool.Run()

	close(results)

	for result := range results {

		out, err := result.Result()

		if err != nil {
			wwlog.Error("%s: %s", result.Ipaddr, out)
			returnErr = err
			continue
		}

		fmt.Printf("%s: %s\n", result.Ipaddr, out)

	}

	return returnErr
}
