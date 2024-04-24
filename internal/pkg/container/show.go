package container

import (
	"fmt"

	"github.com/warewulf/warewulf/internal/pkg/kernel"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
)

type ShowParameter struct {
	Name string
}

type ShowResponse struct {
	Name          string
	Rootfs        string
	Nodes         []string
	KernelVersion string
}

func Show(param *ShowParameter) (*ShowResponse, error) {
	containerName := param.Name

	if !ValidName(containerName) {
		return nil, fmt.Errorf("%s is not a valid container name", containerName)
	}

	rootFsDir := RootFsDir(containerName)
	if !util.IsDir(rootFsDir) {
		return nil, fmt.Errorf("%s is not a valid container", containerName)

	}
	_, kernelVersion, _ := kernel.FindKernel(RootFsDir(containerName))

	nodeDB, err := node.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create nodedb, err: %s", err)
	}

	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		return nil, fmt.Errorf("failed to find all nodes, err: %s", err)
	}

	var nodeList []string
	for _, n := range nodes {
		if n.ContainerName.Get() == containerName {
			nodeList = append(nodeList, n.Id.Get())
		}
	}

	return &ShowResponse{
		Name:          containerName,
		Rootfs:        rootFsDir,
		Nodes:         nodeList,
		KernelVersion: kernelVersion,
	}, nil
}
