package container

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func ValidName(name string) bool {
	if !util.ValidString(name, "^[\\w\\-\\.\\:]+$") {
		wwlog.Warn("VNFS name has illegal characters: %s", name)
		return false
	}
	return true
}

func ListSources() ([]string, error) {
	var ret []string

	err := os.MkdirAll(SourceParentDir(), 0755)
	if err != nil {
		return ret, errors.New("Could not create VNFS source parent directory: " + SourceParentDir())
	}
	wwlog.Debug("Searching for VNFS Rootfs directories: %s", SourceParentDir())

	sources, err := os.ReadDir(SourceParentDir())
	if err != nil {
		return ret, err
	}

	for _, source := range sources {
		wwlog.Verbose("Found VNFS source: %s", source.Name())

		if !ValidName(source.Name()) {
			continue
		}

		if !ValidSource(source.Name()) {
			continue
		}

		ret = append(ret, source.Name())
	}

	return ret, nil
}

func DoesContainerExists(name string) bool {
	fullPath := ImageFile(name)
	return util.IsFile(fullPath)
}

func DoesSourceExist(name string) bool {
	fullPath := RootFsDir(name)
	return util.IsDir(fullPath)
}

func ValidSource(name string) bool {
	if !ValidName(name) {
		return false
	}

	if !DoesSourceExist(name) {
		wwlog.Verbose("Location is not a VNFS source directory: %s", name)
		return false
	}

	return true
}

/*
Delete the chroot of a container
*/
func DeleteSource(name string) error {
	fullPath := SourceDir(name)

	wwlog.Verbose("Removing path: %s", fullPath)
	return os.RemoveAll(fullPath)
}

func Duplicate(name string, destination string) error {
	fullPathImageSource := RootFsDir(name)

	wwlog.Info("Copying sources...")
	err := ImportDirectory(fullPathImageSource, destination)

	if err != nil {
		return err
	}
	wwlog.Info("Building container: %s", destination)
	err = Build(destination, true)
	if err != nil {
		return err
	}
	return nil
}

/*
Delete the image of a container
*/
func DeleteImage(name string) error {
	images_to_delete := []string{
		ImageFile(name),
		ImageFile(name)+".gz",
		ImageFile(name+"-cont"),
		ImageFile(name+"-cont")+".gz",
	}
	for _, imageFile := range images_to_delete {
		if util.IsFile(imageFile) {
			wwlog.Verbose("removing %s for container %s", imageFile, name)
			err := os.Remove(imageFile)
			if err != nil {
				return errors.Errorf("Problem deleting %s for container %s: %s", imageFile, name, err)
			}
		}
	}
	return nil
}

const Includes = "etc/warewulf/includes"
const Excludes = "etc/warewulf/excludes"
const Split = "etc/warewulf/split"

func PatternFile(name string, src string) string {
	return filepath.Join(RootFsDir(name), src)
}

func GetPatterns(name string, src string) ([]string, error) {
	path := PatternFile(name, src)
	if util.IsFile(path) {
		return util.ReadFile(path)
	} else {
		return []string{}, nil
	}
}
