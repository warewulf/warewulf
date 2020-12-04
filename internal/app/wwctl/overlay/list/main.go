package list

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
	"syscall"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	set := make(map[string]int)
	var o []string
	var err error
	var nodeList []node.NodeInfo

	n, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	if SystemOverlay == true {
		if ListLong == false {
			fmt.Printf("%-30s %-12s %-12s\n", "SYSTEM OVERLAY NAME", "NODES", "FILES/DIRS")
		} else {
			fmt.Printf("%-10s %5s %-5s %-18s %s\n", "PERM MODE", "UID", "GID", "SYSTEM-OVERLAY", "FILE PATH")
		}
		o, err = overlay.FindSystemOverlays()
	} else {
		if ListLong == false {
			fmt.Printf("%-30s %-12s %-12s\n", "RUNTIME OVERLAY NAME", "NODES", "FILES/DIRS")
		} else {
			fmt.Printf("%-10s %5s %-5s %-18s %s\n", "PERM MODE", "UID", "GID", "RUNTIME-OVERLAY", "FILE PATH")
		}
		o, err = overlay.FindRuntimeOverlays()
	}
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not get system overlays: %s\n", err)
		return err
	}

	nodeList, err = n.FindAllNodes()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not get node configuration: %s\n", err)
		return err
	}

	for _, node := range nodeList {
		if SystemOverlay == true {
			if node.SystemOverlay.Get() != "" {
				set[node.SystemOverlay.Get()]++
			}
		} else {
			if node.RuntimeOverlay.Get() != "" {
				set[node.RuntimeOverlay.Get()]++
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
			if ListContents == true {
				var fileCount int
				for file := range files {
					fmt.Printf("%-30s %-12d /%-12s\n", name, set[name], files[file])
					fileCount++
				}
				if fileCount == 0 {
					fmt.Printf("%-30s %-12d %-12d\n", name, set[name], 0)
				}
			} else if ListLong == true {
				for file := range files {
					s, err := os.Stat(files[file])
					if err != nil {
						continue
					}

					fileMode := s.Mode()
					perms := fileMode & os.ModePerm

					sys := s.Sys()

					fmt.Printf("%v %5d %-5d %-18s /%s\n", perms, sys.(*syscall.Stat_t).Uid, sys.(*syscall.Stat_t).Gid, o[overlay], files[file])

				}
			} else {
				fmt.Printf("%-30s %-12d %-12d\n", name, set[name], len(files))
			}

		} else {
			wwlog.Printf(wwlog.ERROR, "system/%s (path not found:%s)\n", o[overlay], path)
		}
	}

	var unconfigured bool
	for overlay, _ := range set {
		var overlayPath string

		if SystemOverlay == true {
			overlayPath = config.SystemOverlaySource(overlay)
		} else {
			overlayPath = config.SystemOverlaySource(overlay)
		}

		if util.IsDir(overlayPath) == false {
			fmt.Printf("%-30s %-12d 0\n", "("+overlay+")", set[overlay])
			unconfigured = true
		}
	}

	if unconfigured == true {
		fmt.Printf("\n")
		wwlog.Printf(wwlog.WARN, "There are unconfigured overlays present, run the following command to\n")
		wwlog.Printf(wwlog.WARN, "create a new overlay:\n")
		wwlog.Printf(wwlog.WARN, "\n")
		wwlog.Printf(wwlog.WARN, "   $ sudo wwctl overlay create ...\n")
	}

	return nil
}
