package list

import (
	"fmt"

	"github.com/hpcng/warewulf/internal/pkg/api/container"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {

	containerInfo, err := container.ContainerList()
	if err != nil {
		wwlog.Error("%s", err)
		return
	}

	fmt.Printf("%-25s %-6s %-6s\n", "CONTAINER NAME", "NODES", "KERNEL VERSION")
	for i := 0; i < len(containerInfo); i++ {
		fmt.Printf("%-25s %-6d %-6s\n",
			containerInfo[i].Name,
			containerInfo[i].NodeCount,
			containerInfo[i].KernelVersion)
	}
	return
}
