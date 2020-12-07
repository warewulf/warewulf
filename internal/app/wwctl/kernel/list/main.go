package list

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/kernel"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
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

	fmt.Printf("%-35s %-6s %-6s\n", "VNFS NAME", "BUILT", "NODES")
	for _, k := range kernels {
		image := kernel.KernelImage(k)

		if nodemap[k] == 0 {
			nodemap[k] = 0
		}
		fmt.Printf("%-35s %-6t %-6d\n", k, util.IsFile(image), nodemap[k])

	}

	/*
		nconfig, _ := node.New()
		nodes, _ := nconfig.FindAllNodes()
		nodemap := make(map[string]int)

		for _, n := range nodes {
			nodemap[n.KernelVersion.Get()]++
		}

		images, _ := ioutil.ReadDir(config.KernelParentDir())

		fmt.Printf("%-38s %-16s %-16s %s\n", "KERNEL VERSION", "KERNEL SIZE(k)", "KMODS SIZE(k)", "NODES")
		fmt.Println(strings.Repeat("=", 80))

		for _, file := range images {
			if util.IsDir(path.Join(config.KernelParentDir(), file.Name())) {
				var kernel_size int64
				var kmods_size int64
				if util.IsFile(config.KernelImage(file.Name())) {
					s, _ := os.Stat(config.KernelImage(file.Name()))
					kernel_size = s.Size() / 1024
				}
				if util.IsFile(config.KmodsImage(file.Name())) {
					s, _ := os.Stat(config.KmodsImage(file.Name()))
					kmods_size = s.Size() / 1024
				}

				if nodemap[file.Name()] > 0 {
					fmt.Printf("%-38s %-16d %-16d %d\n", file.Name(), kernel_size, kmods_size, nodemap[file.Name()])
				} else {
					fmt.Printf("%-38s %-16d %-16d %d\n", file.Name(), kernel_size, kmods_size, 0)
				}
			}
		}

	*/

	return nil
}
