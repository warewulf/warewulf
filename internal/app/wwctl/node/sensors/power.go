package sensors

import (
	"fmt"
	"os"

	"github.com/hpcng/warewulf/internal/pkg/batch"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/power"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
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
		wwlog.Printf(wwlog.ERROR, "Cloud not get nodeList: %s\n", err)
		os.Exit(1)
	}

	if len(args) > 0 {
		nodes = node.FilterByName(nodes, args)
	} else {
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

	for _, node := range nodes {

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

		fullFlag := full

		batchpool.Submit(func() {
			if fullFlag {
				ipmiCmd.SensorList()
			} else {
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
			wwlog.Printf(wwlog.ERROR, "%s: %s\n", result.NodeName, out)
			returnErr = err
			continue
		}

		fmt.Printf("%s:\n%s\n", result.NodeName, out)

	}

	return returnErr
}
