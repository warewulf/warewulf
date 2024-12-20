package list

import (
	"fmt"
	"os"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/table"
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

	t := table.New(cmd.OutOrStdout())
	if ListLong {
		t.AddHeader("PERM MODE", "UID", "GID", "SYSTEM-OVERLAY", "FILE PATH", "SITE")
	} else {
		t.AddHeader("OVERLAY NAME", "FILES/DIRS", "SITE")
	}

	for _, name := range overlays {
		overlay_ := overlay.GetOverlay(name)

		if !overlay_.Exists() {
			wwlog.Error("system/%s (path not found:%s)", name, overlay_.Rootfs())
			continue
		}

		files := util.FindFiles(overlay_.Rootfs())

		wwlog.Debug("Iterating overlay rootfs: %s", overlay_.Rootfs())
		if ListLong {
			for file := range files {
				s, err := os.Stat(files[file])
				if err != nil {
					continue
				}

				fileMode := s.Mode()
				perms := fileMode & os.ModePerm

				sys := s.Sys()

				t.AddLine(perms, sys.(*syscall.Stat_t).Uid, sys.(*syscall.Stat_t).Gid, name, files[file], overlay_.IsSiteOverlay())
			}
		} else if ListContents {
			var fileCount int
			for file := range files {
				t.AddLine(name, files[file], overlay_.IsSiteOverlay())
				fileCount++
			}
			if fileCount == 0 {
				t.AddLine(name, 0, overlay_.IsSiteOverlay())
			}
		} else {
			t.AddLine(name, len(files), overlay_.IsSiteOverlay())
		}
	}
	t.Print()

	return nil
}
