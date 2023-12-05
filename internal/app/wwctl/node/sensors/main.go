package sensors

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
		wwlog.Info("No nodes found")
		os.Exit(1)
	}

	batchpool := batch.New(50)
	jobcount := len(nodes)
	results := make(chan power.IPMI, jobcount)

	for _, node := range nodes {
		if node.Ipmi.Ipaddr.IsUnspecified() {
			wwlog.Error("%s: No IPMI IP address", node.Id())
			continue
		}
		var conf node.NodeConf
		conf.GetFrom(n)
		ipmiCmd := power.IPMI{IpmiConf: *conf.Ipmi}
		fullFlag := full
		batchpool.Submit(func() {
			if fullFlag {
				//nolint:errcheck
				ipmiCmd.SensorList()
			} else {
				//nolint:errcheck
				ipmiCmd.SDRList()
			}
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

		fmt.Printf("%s:\n%s\n", result.Ipaddr, out)
	}

	return returnErr
}
