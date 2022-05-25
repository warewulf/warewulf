package container

import (
	"path"
	"strings"

	"github.com/pkg/errors"

	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

func Build(name string, buildForce bool) error {

	rootfsPath := RootFsDir(name)
	imagePath := ImageFile(name)

	if !ValidSource(name) {
		return errors.Errorf("Container does not exist: %s", name)
	}

	if !buildForce {
		wwlog.Debug("Checking if there have been any updates to the VNFS directory")
		if util.PathIsNewer(rootfsPath, imagePath) {
			wwlog.Info("Skipping (VNFS is current)")
			return nil
		}
	}

	excludes_file := path.Join(rootfsPath, "./etc/warewulf/excludes")
	ignore := []string{}

	if util.IsFile(excludes_file) {
		ignore, err := util.ReadFile(excludes_file)
		if err != nil {
			return errors.Wrapf(err, "Failed creating directory: %s", imagePath)
		}

		for i, pattern := range ignore {
			if ( strings.HasPrefix(pattern, "/") ) {
				ignore[i] = pattern[1:]
			}
		}
	}

	err := util.BuildFsImage(
		"VNFS container " + name,
		rootfsPath,
		imagePath,
		[]string{"*"},
		ignore,
		// ignore cross-device files
		true,
		"newc")

	return err
}
