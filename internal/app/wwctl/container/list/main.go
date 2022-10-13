package list

import (
	"fmt"
	"time"

	"github.com/hpcng/warewulf/internal/pkg/api/container"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {

	containerInfo, err := container.ContainerList()
	if err != nil {
		wwlog.Error("%s\n", err)
		return
	}

	fmt.Printf("%-16s %-6s %-16s %-20s %-20s %-8s\n", "CONTAINER NAME", "NODES", "KERNEL VERSION", "CREATION TIME", "MODIFICATION TIME", "SIZE")
	for i := 0; i < len(containerInfo); i++ {
		createTime := time.Unix(int64(containerInfo[i].CreateDate), 0)
		modTime := time.Unix(int64(containerInfo[i].ModDate), 0)
		fmt.Printf("%-16s %-6d %-16s %-20s %-20s %-8s\n",
			containerInfo[i].Name,
			containerInfo[i].NodeCount,
			containerInfo[i].KernelVersion,
			createTime.Format(time.RFC822),
			modTime.Format(time.RFC822),
			util.ByteToString(int64(containerInfo[i].Size)))
	}
	return
}
