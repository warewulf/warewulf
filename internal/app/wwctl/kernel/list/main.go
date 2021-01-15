package list

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/kernel"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	kernels, err := kernel.ListKernels()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	nconfig, _ := node.New()
	nodes, _ := nconfig.FindAllNodes()
	nodemap := make(map[string]int)

	for _, n := range nodes {
		nodemap[n.KernelVersion.Get()]++
	}

	fmt.Printf("%-35s %-6s\n", "VNFS NAME", "NODE#")
	for _, k := range kernels {
		fmt.Printf("%-35s %6d\n", k, nodemap[k])
	}

	return nil
}
