package kernels

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/warewulf/warewulf/internal/app/wwctl/table"
	"github.com/warewulf/warewulf/internal/pkg/image"
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

	var sources []string
	if len(args) == 0 {
		if sources_, err := image.ListSources(); err == nil {
			sources = sources_
		} else {
			return err
		}
	} else {
		sources = args
	}

	t := table.New(cmd.OutOrStdout())
	t.AddHeader("Image", "Kernel", "Version", "Default", "Nodes")
	for _, source := range sources {
		imageKernels := kernel.FindKernels(source)
		defaultKernel := imageKernels.Default()
		for _, kernel_ := range imageKernels {
			isDefault := defaultKernel != nil && defaultKernel == kernel_
			defaultStr := strconv.FormatBool(isDefault)
			nodeCount := kernelNodes[*kernel_]
			if isDefault {
				nodeCount = nodeCount + kernelNodes[kernel.Kernel{ImageName: source, Path: ""}]
			}
			t.AddLine(table.Prep([]string{source, kernel_.Path, kernel_.Version(), defaultStr, strconv.Itoa(nodeCount)})...)
		}
	}
	t.Print()

	return nil
}
