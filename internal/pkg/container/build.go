package container

import (
	"path"

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

	ignore := []string{}
	excludes_file := path.Join(rootfsPath, "./etc/warewulf/excludes")
	if util.IsFile(excludes_file) {
		var err error
		ignore, err = util.ReadFile(excludes_file)
		if err != nil {
			return errors.Wrapf(err, "Failed creating directory: %s", imagePath)
		}
	}

	err := util.BuildFsImage(
		"VNFS container "+name,
		rootfsPath,
		imagePath,
		[]string{"*"},
		ignore,
		// ignore cross-device files
		true,
		"newc")

	return err
}
