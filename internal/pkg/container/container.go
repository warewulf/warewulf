package container

import (
	"strconv"
	"time"

	"github.com/hpcng/warewulf/internal/pkg/util"
)

type ContainerListEntry interface {
	GetHeader() []string
	GetValue() []string
}

type ContainerListResponse struct {
	Containers map[string][]ContainerListEntry `yaml:"Containers" json:"Containers"`
}

type ContainerListSimpleEntry struct {
	Nodes            uint32 `yaml:"Nodes" json:"Nodes"`
	KernelVersion    string `yaml:"KernelVersion" json:"KernelVersion"`
	CreationTime     uint64 `yaml:"CreationTime" json:"CreationTime"`
	ModificationTime uint64 `yaml:"ModificationTime" json:"ModificationTime"`
	Size             uint64 `yaml:"Size" json:"Size"`
}

func (c *ContainerListSimpleEntry) GetHeader() []string {
	return []string{"CONTAINER NAME", "NODES", "KERNEL VERSION", "CREATION TIME", "MODIFICATION TIME", "SIZE"}
}

func (c *ContainerListSimpleEntry) GetValue() []string {
	return []string{strconv.FormatUint(uint64(c.Nodes), 10), c.KernelVersion, time.Unix(int64(c.CreationTime), 0).Format(time.RFC822), time.Unix(int64(c.ModificationTime), 0).Format(time.RFC822), util.ByteToString(int64(c.Size))}
}
