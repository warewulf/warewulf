package list

import (
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/table"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

/*
RunE needs a function of type func(*cobraCommand,[]string) err, but
in order to avoid global variables which mess up testing a function of
the required type is returned
*/
func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var overlays []string

		if len(args) > 0 {
			overlays = args
		} else {
			overlays = overlay.FindOverlays()
		}

		t := table.New(cmd.OutOrStdout())
		locationStr := "SITE"
		if vars.ShowPath {
			locationStr = "PATH"
		}
		if vars.ListLong {
			t.AddHeader("PERM MODE", "UID", "GID", "OVERLAY", "FILE PATH", locationStr, "VARS")
		} else {
			t.AddHeader("OVERLAY NAME", "FILES/DIRS", locationStr)
		}

		for _, name := range overlays {
			overlay_, err := overlay.Get(name)

			if err != nil {
				wwlog.Error("%s:%s", name, err)
				continue
			}

			files := util.FindFiles(overlay_.Rootfs())

			wwlog.Debug("Iterating overlay rootfs: %s", overlay_.Rootfs())
			if vars.ListLong {
				for _, file := range files {
					templateVars := []string{}
					if !strings.HasSuffix(file, "/") {
						templateVars = overlay_.ParseVars(file)
					}
					s, err := os.Stat(overlay_.File(file))
					if err != nil {
						wwlog.Warn("%s: %s: %s", name, file, err)
						continue
					}
					fileMode := s.Mode()
					perms := fileMode & os.ModePerm
					sys := s.Sys()
					locLine := strconv.FormatBool(overlay_.IsSiteOverlay())
					if vars.ShowPath {
						locLine = overlay_.Path()
					}
					t.AddLine(perms, sys.(*syscall.Stat_t).Uid, sys.(*syscall.Stat_t).Gid, name, file, locLine, templateVars)
				}
			} else {
				locLine := strconv.FormatBool(overlay_.IsSiteOverlay())
				if vars.ShowPath {
					locLine = overlay_.Path()
				}
				if vars.ListContents {
					var fileCount int
					for file := range files {
						t.AddLine(name, files[file], locLine)
						fileCount++
					}
					if fileCount == 0 {
						t.AddLine(name, 0, locLine)
					}
				} else {
					t.AddLine(name, len(files), locLine)
				}
			}
		}
		t.Print()

		return nil
	}
}
