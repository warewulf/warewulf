package powercycle

import (
	"os"

	"github.com/hpcng/warewulf/internal/pkg/batch"
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

	if len(args) > 0 {
		nodes = node.FilterByName(nodes, hostlist.Expand(args))
	} else {
		//nolint:errcheck
		cmd.Usage()
		os.Exit(1)
	}

	if len(nodes) == 0 {
		wwlog.Info("No nodes found\n")
		os.Exit(1)
	}

	batchpool := batch.New(50)
	jobcount := len(nodes)
	results := make(chan power.IPMI, jobcount)

	for _, node := range nodes {

		if node.Ipmi.Ipaddr.Get() == "" {
			wwlog.Error("%s: No IPMI IP address", node.Id.Get())
			continue
		}
		var ipmiInterface = "lan"
		if node.Ipmi.Interface.Get() != "" {
			ipmiInterface = node.Ipmi.Interface.Get()
		}
		var ipmiPort = "623"
		if node.Ipmi.Port.Get() != "" {
			ipmiPort = node.Ipmi.Port.Get()
		}
		ipmiCmd := power.IPMI{
			NodeName:  node.Id.Get(),
			HostName:  node.Ipmi.Ipaddr.Get(),
			Port:      ipmiPort,
			User:      node.Ipmi.UserName.Get(),
			Password:  node.Ipmi.Password.Get(),
			Interface: ipmiInterface,
			AuthType:  "MD5",
		}

		batchpool.Submit(func() {
			//nolint:errcheck
			ipmiCmd.PowerCycle()
			results <- ipmiCmd
		})

	}

	batchpool.Run()

	close(results)

	for result := range results {

		out, err := result.Result()

		if err != nil {
			wwlog.Error("%s: %s", result.NodeName, out)
			returnErr = err
			continue
		}

		wwlog.Info("%s: %s\n", result.NodeName, out)

	}

	return returnErr
}
