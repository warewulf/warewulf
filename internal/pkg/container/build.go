package container

import (
	"github.com/pkg/errors"

	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// The default image, which will be used for initial boot, will
// consider files from includes if it exists and ignore files from
// excludes and (if it exists) split.
//
// If split exists, a -cont image will also be generated which will
// consider files from split and ignore from exclude.
func Build(name string, buildForce bool) (err error) {
	if !ValidSource(name) {
		return errors.Errorf("Container does not exist: %s", name)
	}

	err = buildInitImage(name, buildForce)
	if err != nil {
		return err
	}

	if util.IsFile(PatternFile(name, Split)) {
		err = buildContImage(name, buildForce)
		if err != nil {
			return err
		}
	}

	return nil
}

func buildInitImage(name string, buildForce bool) (err error) {
	rootfsPath := RootFsDir(name)
	imagePath := ImageFile(name)

	if !buildForce {
		wwlog.Debug("Checking if there have been any updates to the VNFS directory")
		if util.PathIsNewer(rootfsPath, imagePath) {
			wwlog.Info("Skipping %s (VNFS is current)", imagePath)
			return nil
		}
	}

	var includePatterns []string
	includePatterns, err = GetPatterns(name, Includes)
	if err != nil {
		return err
	}
	if len(includePatterns) == 0 {
		includePatterns = append(includePatterns, "/")
	}

	var ignorePatterns []string
	ignorePatterns, err = GetPatterns(name, Excludes)
	if err != nil {
		return err
	}

	var splitPatterns []string
	splitPatterns, err = GetPatterns(name, Split)
	if err != nil {
		return err
	}
	for _, pattern := range splitPatterns {
		ignorePatterns = append(ignorePatterns, pattern)
	}

	return util.BuildFsImage(
		"VNFS container "+name,
		rootfsPath,
		imagePath,
		includePatterns,
		ignorePatterns,
		true, // ignore cross-device files
		"newc")
}

func buildContImage (name string, buildForce bool) (err error) {
	rootfsPath := RootFsDir(name)
	imagePath := ImageFile(name+"-cont")

	if !buildForce {
		wwlog.Debug("Checking if there have been any updates to the VNFS directory")
		if util.PathIsNewer(rootfsPath, imagePath) {
			wwlog.Info("Skipping %s (VNFS is current)", imagePath)
			return nil
		}
	}

	var includePatterns []string
	includePatterns, err = GetPatterns(name, Split)
	if err != nil {
		return err
	}
	if len(includePatterns) == 0 {
		includePatterns = append(includePatterns, "/")
	}

	var ignorePatterns []string
	ignorePatterns, err = GetPatterns(name, Excludes)
	if err != nil {
		return err
	}

	return util.BuildFsImage(
		"VNFS container "+name+"-cont",
		rootfsPath,
		imagePath,
		includePatterns,
		ignorePatterns,
		true, // ignore cross-device files
		"newc")
}
