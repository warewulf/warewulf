package list

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	sources, err := container.ListSources()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	nconfig, _ := node.New()
	nodes, _ := nconfig.FindAllNodes()
	nodemap := make(map[string]int)

	for _, n := range nodes {
		nodemap[n.ContainerName.Get()]++
	}

	fmt.Printf("%-35s %-6s %-6s\n", "VNFS NAME", "BUILT", "NODES")
	for _, source := range sources {
		image := container.ImageFile(source)

		if nodemap[source] == 0 {
			nodemap[source] = 0
		}
		fmt.Printf("%-35s %-6t %-6d\n", source, util.IsFile(image), nodemap[source])

	}

	/*
		nconfig, _ := node.New()
		nodes, _ := nconfig.FindAllNodes()
		nodemap := make(map[string]int)

		for _, n := range nodes {
			nodemap[n.Vnfs.Get()]++
		}

		images, _ := ioutil.ReadDir(config.VnfsImageParentDir())

		fmt.Printf("%-38s %-16s %s\n", "VNFS Name", "VNFS SIZE(k)", "NODES")
		fmt.Println(strings.Repeat("=", 80))

		for _, file := range images {
			v, err := container.Load(file.Name())
			if err == nil {
				var vnfs_size int64
				if util.IsFile(config.VnfsImage(file.Name())) {
					s, _ := os.Stat(config.VnfsImage(file.Name()))
					vnfs_size = s.Size() / 1024
				}

				if nodemap[v.Source] > 0 {
					fmt.Printf("%-38s %-16d %d\n", v.Source, vnfs_size, nodemap[v.Source])
				} else {
					fmt.Printf("%-38s %-16d %d\n", v.Source, vnfs_size, 0)
				}

			}
			continue

			if util.IsDir(path.Join(config.VnfsImageParentDir(), file.Name())) {
				var vnfs_size int64
				if util.IsFile(config.VnfsImage(file.Name())) {
					s, _ := os.Stat(config.VnfsImage(file.Name()))
					vnfs_size = s.Size() / 1024
				}

				if nodemap[file.Name()] > 0 {
					fmt.Printf("%-38s %-16d %d\n", file.Name(), vnfs_size, nodemap[file.Name()])
				} else {
					fmt.Printf("%-38s %-16d %d\n", file.Name(), vnfs_size, 0)
				}
			}
		}
	*/

	return nil
}
