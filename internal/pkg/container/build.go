package container

import (
	"fmt"
	"path"

	"github.com/pkg/errors"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
	"golang.org/x/sync/errgroup"
)

type BuildParameter struct {
	Names   []string
	Force   bool
	Default bool
}

func Build(param *BuildParameter) error {
	if len(param.Names) == 0 {
		return fmt.Errorf("build names array is empty")
	}

	if param.Default && len(param.Names) != 1 {
		return fmt.Errorf("can only set default for one container")
	}

	g := new(errgroup.Group)

	for _, name := range param.Names {
		if !ValidSource(name) {
			return fmt.Errorf("build source %s does not exist", name)
		}

		n := name
		g.Go(func() error {
			err := build(n, param.Force)
			return err
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("error occurs during build, err: %s", err)
	}

	if param.Default {
		return SetProfileDefaultContainer(param.Names[0])
	}

	return nil
}

func build(name string, buildForce bool) error {

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
