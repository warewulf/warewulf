package ready

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/vnfs"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	n, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	nodes, err := n.FindAllNodes()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Cloud not get nodeList: %s\n", err)
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

		if node.Vnfs.Get() != "" {
			v, _ := vnfs.Load(node.Vnfs.Get())
			if util.IsFile(v.Image) == true {
				vnfs_good = true
			} else {
				status = false
				wwlog.Printf(wwlog.VERBOSE, "VNFS not found: %s, %s\n", node.Id.Get(), v.Source)
			}
		} else {
			status = false
			wwlog.Printf(wwlog.VERBOSE, "Node Kernel not defined: %s\n", node.Id.Get())
		}

		if node.KernelVersion.Get() != "" {
			if util.IsFile(config.KernelImage(node.KernelVersion.Get())) == true {
				kernel_good = true
			} else {
				status = false
				wwlog.Printf(wwlog.VERBOSE, "Node Kernel not found: %s, %s\n", node.Id.Get(), node.KernelVersion.Get())
			}
			if util.IsFile(config.KmodsImage(node.KernelVersion.Get())) == true {
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
			if util.IsFile(config.SystemOverlayImage(node.Id.Get())) == true {
				systemo_good = true
			} else {
				status = false
				wwlog.Printf(wwlog.VERBOSE, "System Overlay not found: %s\n", config.SystemOverlayImage(node.Id.Get()))
			}
		} else {
			status = false
			wwlog.Printf(wwlog.VERBOSE, "System Overlay not defined: %s\n", node.Id.Get())
		}

		if node.RuntimeOverlay.Get() != "" {
			if util.IsFile(config.RuntimeOverlayImage(node.Id.Get())) == true {
				runtimeo_good = true
			} else {
				status = false
				wwlog.Printf(wwlog.VERBOSE, "Runtime Overlay not found: %s\n", config.RuntimeOverlaySource(node.Id.Get()))
			}
		} else {
			status = false
			wwlog.Printf(wwlog.VERBOSE, "Runtime Overlay not defined: %s\n", node.Id.Get())
		}

		fmt.Printf("%-25s %-10t %-6t %-6t %-6t %-6t %-6t\n", node.Id.Get(), status, vnfs_good, kernel_good, kmods_good, systemo_good, runtimeo_good)
	}

	return nil
}
