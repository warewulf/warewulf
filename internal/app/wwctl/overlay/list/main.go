package list

import (
	"strconv"

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
		t.AddHeader("OVERLAY NAME", "FILES/DIRS", locationStr)

		for _, name := range overlays {
			overlay_, err := overlay.Get(name)

			if err != nil {
				wwlog.Error("%s:%s", name, err)
				continue
			}

			files := util.FindFiles(overlay_.Rootfs())

			wwlog.Debug("Iterating overlay rootfs: %s", overlay_.Rootfs())
			locLine := strconv.FormatBool(overlay_.IsSiteOverlay())
			if vars.ShowPath {
				locLine = overlay_.Path()
			}
			if vars.ListContents || vars.ListLong {
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
		t.Print()

		return nil
	}
}
