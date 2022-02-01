package list

import (
	"fmt"

	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	containerName := args[1]
	if !container.ValidName(containerName) {
		return fmt.Errorf("%s is not a valid container", containerName)
	}
	if !ShowAll {
		fmt.Printf("%s\n", container.RootFsDir(containerName))
	} else {
		fmt.Printf("Name: %s\n", containerName)
		fmt.Printf("Rootsfs: %s\n", container.RootFsDir(containerName))
		nodeDB, _ := node.New()

		nodes, _ := nodeDB.FindAllNodes()
		var nodeList []string
		for _, n := range nodes {
			if n.ContainerName.Get() == containerName {

				nodeList = append(nodeList, n.Id.Get())
			}
		}
		fmt.Printf("Nr nodes: %d\n", len(nodeList))
		fmt.Printf("Nodes: %s\n", nodeList)

	}
	return nil
}
