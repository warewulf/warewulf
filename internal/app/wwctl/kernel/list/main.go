package list

import (
	"fmt"
	"os"

	"github.com/hpcng/warewulf/internal/pkg/kernel"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	kernels, err := kernel.ListKernels()
	if err != nil {
		wwlog.Error("%s", err)
		os.Exit(1)
	}

	nconfig, _ := node.New()
	nodes, _ := nconfig.FindAllNodes()
	nodemap := make(map[string]int)

	for _, n := range nodes {
		nodemap[n.Kernel.Override.Get()]++
	}

	fmt.Printf("%-35s %-25s %-6s\n", "KERNEL NAME", "KERNEL VERSION", "NODES")
	for _, k := range kernels {
		fmt.Printf("%-35s %-25s %6d\n", k, kernel.GetKernelVersion(k), nodemap[k])
	}

	return nil
}
