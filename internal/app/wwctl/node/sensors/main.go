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
		nodes = node.FilterByName(nodes, args)
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
			wwlog.Error("%s: %s", result.NodeName, out)
			returnErr = err
			continue
		}

		fmt.Printf("%s:\n%s\n", result.NodeName, out)
	}

	return returnErr
}
