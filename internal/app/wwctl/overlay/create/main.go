package create

import (
	"github.com/hpcng/warewulf/internal/pkg/assets"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
)



func CobraRunE(cmd *cobra.Command, args []string) error {

	if len(args) < 1 {
		cmd.Help()
		os.Exit(1)
	}

	if SystemOverlay == true {
		err := overlay.SystemOverlayInit(args[0])
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}
		wwlog.Printf(wwlog.INFO, "Created new system overlay: %s\n", args[0])
	} else {
		err := overlay.RuntimeOverlayInit(args[0])
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}
		wwlog.Printf(wwlog.INFO, "Created new runtime overlay: %s\n", args[0])
	}

	if NoOverlayUpdate == false {
		nodes, err := assets.FindAllNodes()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Cloud not get nodeList: %s\n", err)
			os.Exit(1)
		}

		var updateNodes []assets.NodeInfo

		for _, node := range nodes {
			if SystemOverlay == true && node.SystemOverlay == args[0] {
				updateNodes = append(updateNodes, node)
			} else if node.RuntimeOverlay == args[0] {
				updateNodes = append(updateNodes, node)
			}
		}

		if SystemOverlay == true {
			wwlog.Printf(wwlog.INFO, "Updating System Overlays...\n")
			return overlay.SystemBuild(updateNodes, true)
		} else {
			wwlog.Printf(wwlog.INFO, "Updating Runtime Overlays...\n")
			return overlay.RuntimeBuild(updateNodes, true)
		}
	}


	return nil
}