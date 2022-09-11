package list

import (
	"fmt"
	"os"
	"syscall"

	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var overlays []string

	if len(args) > 0 {
		overlays = args
	} else {
		var err error
		overlays, err = overlay.FindOverlays()
		if err != nil {
			return errors.Wrap(err, "could not obtain list of overlays from system")
		}
	}

	if ListLong {
		fmt.Printf("%-10s %5s %-5s %-18s %s\n", "PERM MODE", "UID", "GID", "SYSTEM-OVERLAY", "FILE PATH")
	} else {
		fmt.Printf("%-30s %-12s\n", "OVERLAY NAME", "FILES/DIRS")
	}

	for o := range overlays {
		name := overlays[o]
		path := overlay.OverlaySourceDir(name)

		if util.IsDir(path) {
			files := util.FindFiles(path)

			wwlog.Debug("Iterating overlay path: %s\n", path)
			if ListContents {
				var fileCount int
				for file := range files {
					fmt.Printf("%-30s /%-12s\n", name, files[file])
					fileCount++
				}
				if fileCount == 0 {
					fmt.Printf("%-30s %-12d\n", name, 0)
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

					fmt.Printf("%v %5d %-5d %-18s /%s\n", perms, sys.(*syscall.Stat_t).Uid, sys.(*syscall.Stat_t).Gid, overlays[o], files[file])

				}
			} else {
				fmt.Printf("%-30s %-12d\n", name, len(files))
			}

		} else {
			wwlog.Error("system/%s (path not found:%s)\n", overlays[o], path)
		}
	}

	return nil
}
