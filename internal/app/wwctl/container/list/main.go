package list

import (
	"fmt"
	"os"
	"path"

	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	sources, err := container.ListSources()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	nodeDB, _ := node.New()
	nodes, _ := nodeDB.FindAllNodes()
	nodemap := make(map[string]int)

	for _, n := range nodes {
		nodemap[n.ContainerName.Get()]++
	}

	fmt.Printf("%-25s %-6s %-6s\n", "CONTAINER NAME", "NODES", "KERNEL")
	for _, source := range sources {
		if nodemap[source] == 0 {
			nodemap[source] = 0
		}

		kernel := container.KernelFind(source)
		fmt.Printf("%-25s %-6d %s\n", source, nodemap[source], path.Base(kernel))

	}
	return nil
}
