package print_mnts

import (
	"fmt"

	"github.com/hpcng/warewulf/internal/pkg/container"
	"gopkg.in/yaml.v2"
)

func main() {
	mounts := []container.MntDetails{
		container.MntDetails{
			Source:   "/etc/resolv.conf",
			Dest:     "/etc/resolv.conf",
			ReadOnly: false,
		},
		container.MntDetails{
			Source:   "etc/zypp/credentials.d/SCCcredentials",
			Dest:     "etc/zypp/credentials.d/SCCcredentials",
			ReadOnly: false,
		},
		container.MntDetails{
			Source:   "/etc/SUSEConnect",
			Dest:     "/etc/SUSEConnect",
			ReadOnly: false,
		},
	}
	var ptrs []*container.MntDetails
	for _, m := range mounts {
		ptrs = append(ptrs, &m)
	}
	mountPts := container.MountPoints{
		MountPnts: ptrs,
	}
	data, _ := yaml.Marshal(mountPts)
	fmt.Println(data)
}
