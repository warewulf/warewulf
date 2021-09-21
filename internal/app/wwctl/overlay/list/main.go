package list

import (
	"fmt"
	"os"
	"syscall"

	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	set := make(map[string]int)
	var o []string
	var err error
	var nodeList []node.NodeInfo
	var overlayName string

	if len(args) < 1 {
		return errors.New("overlay kind must be specified. Use -h or --help for additional information.")
	}

	overlayKind := args[0]

	if overlayKind != "system" && overlayKind != "runtime" {
		return errors.New("overlay kind must be of type 'system' or 'runtime'")
	}

	if len(args) > 1 {
		overlayName = args[1]
	}

	n, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	if overlayKind == "system" {
		if !ListLong {
			fmt.Printf("%-30s %-12s %-12s\n", "SYSTEM OVERLAY NAME", "NODES", "FILES/DIRS")
		} else {
			fmt.Printf("%-10s %5s %-5s %-18s %s\n", "PERM MODE", "UID", "GID", "SYSTEM-OVERLAY", "FILE PATH")
		}
		o, err = overlay.FindSystemOverlays()
	} else if overlayKind == "runtime" {
		if !ListLong {
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
		if overlayKind == "system" {
			if node.SystemOverlay.Get() != "" {
				set[node.SystemOverlay.Get()]++
			}
		} else if overlayKind == "runtime" {
			if node.RuntimeOverlay.Get() != "" {
				set[node.RuntimeOverlay.Get()]++
			}
		}
	}

	for overlay := range o {
		var path string
		name := o[overlay]

		if overlayName != "" && overlayName != name {
			continue
		}

		if overlayKind == "system" {
			path = config.SystemOverlaySource(o[overlay])
		} else if overlayKind == "runtime" {
			path = config.RuntimeOverlaySource(o[overlay])
		}

		if util.IsDir(path) {
			files := util.FindFiles(path)

			wwlog.Printf(wwlog.DEBUG, "Iterating overlay path: %s\n", path)
			if ListContents {
				var fileCount int
				for file := range files {
					fmt.Printf("%-30s %-12d /%-12s\n", name, set[name], files[file])
					fileCount++
				}
				if fileCount == 0 {
					fmt.Printf("%-30s %-12d %-12d\n", name, set[name], 0)
				}
			} else if ListLong {
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

	return nil
}
