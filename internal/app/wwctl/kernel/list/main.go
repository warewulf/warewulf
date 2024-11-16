package list

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/warewulf/warewulf/internal/app/wwctl/table"
	"github.com/warewulf/warewulf/internal/pkg/container"
	"github.com/warewulf/warewulf/internal/pkg/kernel"
	"github.com/warewulf/warewulf/internal/pkg/node"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	nodesYaml, err := node.New()
	if err != nil {
		return err
	}
	nodes, err := nodesYaml.FindAllNodes()
	if err != nil {
		return err
	}
	kernelNodes := make(map[kernel.Kernel]int)
	for _, node := range nodes {
		kernel_ := kernel.FromNode(&node)
		if kernel_ != nil {
			kernelNodes[*kernel_]++
		}
	}

	sources, err := container.ListSources()
	if err != nil {
		return err
	}

	t := table.New(cmd.OutOrStdout())
	t.AddHeader("Container", "Kernel", "Version", "Preferred", "Nodes")
	for _, source := range sources {
		containerKernels := kernel.FindKernels(source)
		preferredKernel := containerKernels.Preferred()
		for _, kernel_ := range containerKernels {
			preferred := preferredKernel != nil && preferredKernel == kernel_
			preferredStr := strconv.FormatBool(preferred)
			nodeCount := kernelNodes[*kernel_]
			if preferred {
				nodeCount = nodeCount + kernelNodes[kernel.Kernel{ContainerName: source, Path: ""}]
			}
			t.AddLine(table.Prep([]string{source, kernel_.Path, kernel_.Version(), preferredStr, strconv.Itoa(nodeCount)})...)
		}
	}
	t.Print()

	return nil
}
