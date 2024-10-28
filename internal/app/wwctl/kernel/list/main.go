package list

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/kernel"
	"github.com/warewulf/warewulf/internal/pkg/node"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	kernels, err := kernel.ListKernels()
	if err != nil {
		return err
	}

	nconfig, _ := node.New()
	nodes, _ := nconfig.FindAllNodes()
	nodemap := make(map[string]int)

	for _, n := range nodes {
		nodemap[n.Kernel.Override]++
	}

	fmt.Printf("%-35s %-25s %-6s\n", "KERNEL NAME", "KERNEL VERSION", "NODES")
	for _, k := range kernels {
		fmt.Printf("%-35s %-25s %6d\n", k, kernel.GetKernelVersion(k), nodemap[k])
	}

	return nil
}
