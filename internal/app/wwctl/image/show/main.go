package show

import (
	"fmt"

	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/kernel"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"

	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	imageName := args[0]

	if !image.ValidName(imageName) {
		return fmt.Errorf("%s is not a valid image name", imageName)
	}

	rootFsDir := image.RootFsDir(imageName)
	if !util.IsDir(rootFsDir) {
		return fmt.Errorf("%s is not a valid image", imageName)
	}
	kernel := kernel.FindKernels(imageName).Default()
	kernelVersion := ""
	if kernel != nil {
		kernelVersion = kernel.Version()
	}

	nodeDB, err := node.New()
	if err != nil {
		return err
	}

	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		return err
	}

	var nodeList []string
	for _, n := range nodes {
		if n.ImageName == imageName {
			nodeList = append(nodeList, n.Id())
		}
	}

	if !ShowAll {
		fmt.Printf("%s\n", rootFsDir)
	} else {
		if kernelVersion == "" {
			kernelVersion = "not found"
		}
		fmt.Printf("Name: %s\n", imageName)
		fmt.Printf("KernelVersion: %s\n", kernelVersion)
		fmt.Printf("Rootfs: %s\n", rootFsDir)
		fmt.Printf("Nr nodes: %d\n", len(nodeList))
		fmt.Printf("Nodes: %v\n", nodeList)
	}

	return nil
}
