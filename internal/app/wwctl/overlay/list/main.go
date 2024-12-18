package list

import (
	"fmt"
	"os"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var overlays []string

	if len(args) > 0 {
		overlays = args
	} else {
		var err error
		overlays, err = overlay.FindOverlays()
		if err != nil {
			return fmt.Errorf("could not obtain list of overlays from system: %w", err)
		}
	}

	if ListLong {
		wwlog.Info("%-10s %5s %-5s %-18s %s\n", "PERM MODE", "UID", "GID", "SYSTEM-OVERLAY", "FILE PATH", "SITE")
	} else {
		wwlog.Info("%-30s %-12s-%12s\n", "OVERLAY NAME", "FILES/DIRS", "SITE")
	}

	for o := range overlays {
		name := overlays[o]
		path, isSite := overlay.OverlaySourceDir(name)

		if util.IsDir(path) {
			files := util.FindFiles(path)

			wwlog.Debug("Iterating overlay path: %s", path)
			if ListLong {
				for file := range files {
					s, err := os.Stat(files[file])
					if err != nil {
						continue
					}

					fileMode := s.Mode()
					perms := fileMode & os.ModePerm

					sys := s.Sys()

					wwlog.Info("%v %5d %-5d %-18s /%s\n", perms, sys.(*syscall.Stat_t).Uid, sys.(*syscall.Stat_t).Gid, overlays[o], files[file], isSite)
				}
			} else if ListContents {
				var fileCount int
				for file := range files {
					wwlog.Info("%-30s /%-12s\n", name, files[file])
					fileCount++
				}
				if fileCount == 0 {
					wwlog.Info("%-30s %-12d\n", name, 0)
				}
			} else {
				wwlog.Info("%-30s %-12d\n", name, len(files), isSite)
			}

		} else {
			wwlog.Error("system/%s (path not found:%s)", overlays[o], path)
		}
	}

	return nil
}
