package image

import (
	"fmt"
	"path"

	"github.com/pkg/errors"

	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Build(name string, buildForce bool) error {
	wwlog.Info("Building image: %s", name)

	rootfsPath := RootFsDir(name)
	imagePath := ImageFile(name)

	if !ValidSource(name) {
		return errors.Errorf("Image does not exist: %s", name)
	}

	if !buildForce {
		wwlog.Debug("Checking if there have been any updates to the image source directory")
		if util.PathIsNewer(rootfsPath, imagePath) {
			wwlog.Info("Skipping (Image is current)")
			return nil
		}
	}

	ignore := []string{}
	excludes_file := path.Join(rootfsPath, "./etc/warewulf/excludes")
	if util.IsFile(excludes_file) {
		var err error
		ignore, err = util.ReadFile(excludes_file)
		if err != nil {
			return fmt.Errorf("failed creating directory: %s: %w", imagePath, err)
		}
	}

	err := util.BuildFsImage(
		"Image "+name,
		rootfsPath,
		imagePath,
		[]string{"*"},
		ignore,
		// ignore cross-device files
		true,
		"newc",
		// cpio args
		"--renumber-inodes")

	return err
}
