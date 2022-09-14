package container

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"

	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

func ValidName(name string) bool {
	if !util.ValidString(name, "^[\\w\\-\\.\\:]+$") {
		wwlog.Warn("VNFS name has illegal characters: %s\n", name)
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
	wwlog.Debug("Searching for VNFS Rootfs directories: %s\n", SourceParentDir())

	sources, err := ioutil.ReadDir(SourceParentDir())
	if err != nil {
		return ret, err
	}

	for _, source := range sources {
		wwlog.Verbose("Found VNFS source: %s\n", source.Name())

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

func ValidSource(name string) bool {
	fullPath := RootFsDir(name)

	if !ValidName(name) {
		return false
	}

	if !util.IsDir(fullPath) {
		wwlog.Verbose("Location is not a VNFS source directory: %s\n", name)
		return false
	}

	return true
}

/*
Delete the chroot of a container
*/
func DeleteSource(name string) error {
	fullPath := SourceDir(name)

	wwlog.Verbose("Removing path: %s\n", fullPath)
	return os.RemoveAll(fullPath)
}

/*
Delete the image of a container
*/
func DeleteImage(name string) error {
	imageFile := ImageFile(name)
	if util.IsFile(imageFile) {
		wwlog.Verbose("removing %s for container %s\n", imageFile, name)
		errImg := os.Remove(imageFile)
		wwlog.Verbose("removing %s for container %s\n", imageFile+".gz", name)
		errGz := os.Remove(imageFile + ".gz")
		if errImg != nil {
			return errors.Errorf("Problems delete %s for container %s: %s\n", imageFile, name, errImg)
		}
		if errGz != nil {
			return errors.Errorf("Problems delete %s for container %s: %s\n", imageFile+".gz", name, errGz)
		}
		return nil
	}
	return errors.Errorf("Image %s of container %s doesn't exist\n", imageFile, name)
}
