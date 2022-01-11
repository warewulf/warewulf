package ready

import (
	"fmt"
	"os"

	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/kernel"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	n, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	nodes, err := n.FindAllNodes()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not get node list: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("%-25s %-10s %-6s %-6s %-6s %-6s %-6s\n", "NODE NAME", "STATUS", "VNFS", "KERNEL", "KMODS", "SYS-OL", "RUN-OL")

	for _, node := range nodes {
		var vnfs_good bool
		var kernel_good bool
		var kmods_good bool
		var systemo_good bool
		var runtimeo_good bool
		status := true

		if node.ContainerName.Get() != "" {
			vnfsImage := container.ImageFile(node.ContainerName.Get())

			if util.IsFile(vnfsImage) {
				vnfs_good = true
			} else {
				status = false
				wwlog.Printf(wwlog.VERBOSE, "VNFS not found: %s, %s\n", node.Id.Get(), vnfsImage)
			}
		} else {
			status = false
			wwlog.Printf(wwlog.VERBOSE, "Node Kernel not defined: %s\n", node.Id.Get())
		}

		if node.KernelVersion.Get() != "" {
			if util.IsFile(kernel.KernelImage(node.KernelVersion.Get())) {
				kernel_good = true
			} else {
				status = false
				wwlog.Printf(wwlog.VERBOSE, "Node Kernel not found: %s, %s\n", node.Id.Get(), node.KernelVersion.Get())
			}
			if util.IsFile(kernel.KmodsImage(node.KernelVersion.Get())) {
				kmods_good = true
			} else {
				status = false
				wwlog.Printf(wwlog.VERBOSE, "Node Kmods not found: %s, %s\n", node.Id.Get(), node.KernelVersion.Get())
			}
		} else {
			status = false
			wwlog.Printf(wwlog.VERBOSE, "Node Kernel version not defined: %s\n", node.Id.Get())
		}

		if node.SystemOverlay.Get() != "" {
			if util.IsFile(overlay.SystemOverlayImage(node.Id.Get())) {
				systemo_good = true
			} else {
				status = false
				wwlog.Printf(wwlog.VERBOSE, "System Overlay not found: %s\n", overlay.SystemOverlayImage(node.Id.Get()))
			}
		} else {
			status = false
			wwlog.Printf(wwlog.VERBOSE, "System Overlay not defined: %s\n", node.Id.Get())
		}

		fmt.Printf("%-25s %-10t %-6t %-6t %-6t %-6t %-6t\n", node.Id.Get(), status, vnfs_good, kernel_good, kmods_good, systemo_good, runtimeo_good)
	}

	return nil
}
