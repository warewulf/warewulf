package imprt

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	var dest string

	source := args[1]

	if len(args) == 3 {
		dest = args[2]
	} else {
		dest = source
	}
	overlay_, err := overlay.GetOverlay(args[0])
	if err != nil {
		return err
	}
	if !overlay_.IsSiteOverlay() {
		overlay_, err = overlay_.CloneSiteOverlay()
		if err != nil {
			return err
		}
	}

	if util.IsDir(overlay_.File(dest)) {
		dest = path.Join(dest, path.Base(source))
	}

	if !OverwriteFile && util.IsFile(overlay_.File(dest)) {
		return fmt.Errorf("a file with that name already exists in the overlay")
	}

	if CreateDirs {
		parent := filepath.Dir(overlay_.File(dest))
		if _, err = os.Stat(parent); os.IsNotExist(err) {
			wwlog.Debug("Create dir: %s", parent)
			srcInfo, err := os.Stat(source)
			if err != nil {
				return fmt.Errorf("could not retrieve the stat for file: %w", err)
			}
			mode := srcInfo.Mode()
			mode |= ((mode & 0444) >> 2) // add execute permission wherever srcInfo has read
			err = os.MkdirAll(parent, mode)
			if err != nil {
				return fmt.Errorf("could not create parent dir: %s: %w", parent, err)
			}
		}
	}

	err = util.CopyFile(source, overlay_.File(dest))
	if err != nil {
		return fmt.Errorf("could not copy file into overlay: %w", err)
	}

	return nil
}
