package poweroff

import (
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/power"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/hpcng/warewulf/internal/pkg/batch"
	"github.com/spf13/cobra"
	"os"
	"fmt"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var returnErr error = nil
	var nodeList []node.NodeInfo

	n, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	if len(args) >= 1 {
		nodeList, _ = n.SearchByNameList(args)
	} else {
		wwlog.Printf(wwlog.ERROR, "No requested nodes\n")
		os.Exit(255)
	}

	if len(nodeList) == 0 {
		wwlog.Printf(wwlog.ERROR, "No nodes found matching: '%s'\n", args[0])
		os.Exit(255)
	}

	batchpool := batch.New(50, 0)
	jobcount := len(nodeList)
	results := make(chan power.IPMI, jobcount)

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

		batchpool.Submit(func() {
			ipmiCmd.PowerOff()
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

		fmt.Printf("%s: %s\n", result.NodeName, out)

	}

	return returnErr
}

