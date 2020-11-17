package list

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/assets"
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	config := config.New()
	set := make(map[string]int)
	var o []string
	var err error
	var nodeList []assets.NodeInfo

	if SystemOverlay == true {
		fmt.Printf("%-25s %-8s %-8s\n", "SYSTEM OVERLAY NAME", "NODES", "FILES")
		o, err = overlay.FindAllSystemOverlays()
	} else {
		fmt.Printf("%-25s %-8s %-8s\n", "RUNTIME OVERLAY NAME", "NODES", "FILES")
		o, err = overlay.FindAllRuntimeOverlays()
	}
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not get system overlays: %s\n", err)
		return err
	}

	nodeList, err = assets.FindAllNodes()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not get node configuration: %s\n", err)
		return err
	}

	for _, node := range nodeList {
		if SystemOverlay == true {
			if node.SystemOverlay != "" {
				set[node.SystemOverlay] ++
			}
		} else {
			if node.RuntimeOverlay != "" {
				set[node.RuntimeOverlay] ++
			}
		}
	}

	for overlay := range o {
		var path string
		name := o[overlay]

		if len(args) > 0 {
			if args[0] != name {
				continue
			}
		}

		if SystemOverlay == true {
			path = config.SystemOverlaySource(o[overlay])
		} else {
			path = config.RuntimeOverlaySource(o[overlay])
		}

		if util.IsDir(path) == true {
			files := util.FindFiles(path)

			wwlog.Printf(wwlog.DEBUG, "Iterating overlay path: %s\n", path)
			if ListFiles == true {
				var fileCount int
				for file := range files {
					fmt.Printf("%-25s %-8d /%s\n", name, set[name], files[file])
					fileCount ++
				}
				if fileCount == 0 {
					fmt.Printf("%-25s %-8d %-8d\n", name, set[name], 0)
				}
			} else {
				fmt.Printf("%-25s %-8d %-8d\n", name, set[name], len(files))
			}

		} else {
			wwlog.Printf(wwlog.ERROR, "system/%s (path not found:%s)\n", o[overlay], path)
		}
	}

	return nil
}
