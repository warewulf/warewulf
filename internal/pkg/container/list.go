package container

import (
	"fmt"
	"os"

	"github.com/warewulf/warewulf/internal/pkg/kernel"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

type ListResponse struct {
	Name          string
	NodeCount     uint32
	KernelVersion string
	CreateDate    uint64
	ModDate       uint64
	Size          uint64
}

func List() ([]*ListResponse, error) {
	var sources []string

	sources, err := ListSources()
	if err != nil {
		return nil, fmt.Errorf("failed to list all sources, err: %s", err)
	}

	nodeDB, err := node.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create nodedb, err: %s", err)
	}

	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		return nil, fmt.Errorf("failed to find all nodes, err: %s", err)
	}

	nodemap := make(map[string]int)
	for _, n := range nodes {
		nodemap[n.ContainerName.Get()]++
	}

	var ret []*ListResponse
	for _, source := range sources {
		if nodemap[source] == 0 {
			nodemap[source] = 0
		}

		wwlog.Debug("Finding kernel version for: %s", source)
		_, kernelVersion, _ := kernel.FindKernel(RootFsDir(source))
		var creationTime uint64
		sourceStat, err := os.Stat(SourceDir(source))
		if err != nil {
			wwlog.Error("%s\n", err)
		} else {
			creationTime = uint64(sourceStat.ModTime().Unix())
		}
		var modTime uint64
		imageStat, err := os.Stat(ImageFile(source))
		if err == nil {
			modTime = uint64(imageStat.ModTime().Unix())
		}
		size, err := util.DirSize(SourceDir(source))
		if err != nil {
			wwlog.Error("%s\n", err)
		}
		imgSize, err := os.Stat(ImageFile(source))
		if err == nil {
			size += imgSize.Size()
		}
		imgSize, err = os.Stat(ImageFile(source) + ".gz")
		if err == nil {
			size += imgSize.Size()
		}

		ret = append(ret, &ListResponse{
			Name:          source,
			NodeCount:     uint32(nodemap[source]),
			KernelVersion: kernelVersion,
			CreateDate:    creationTime,
			ModDate:       modTime,
			Size:          uint64(size),
		})

	}
	return ret, nil
}
