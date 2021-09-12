package list

import (
	"fmt"
	"os"

	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
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

	fmt.Printf("%-35s %-6s %-6s\n", "CONTAINER NAME", "BUILT", "NODES")
	for _, source := range sources {
		image := container.ImageFile(source)

		if nodemap[source] == 0 {
			nodemap[source] = 0
		}
		fmt.Printf("%-35s %-6t %-6d\n", source, util.IsFile(image), nodemap[source])

	}
	return nil
}
