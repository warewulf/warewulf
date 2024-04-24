package container

import (
	"fmt"
	"os"

	"github.com/pkg/errors"

	"github.com/warewulf/warewulf/internal/pkg/node"
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

/*
Delete the image of a container
*/
func DeleteImage(name string) error {
	imageFile := ImageFile(name)
	if util.IsFile(imageFile) {
		wwlog.Verbose("removing %s for container %s", imageFile, name)
		errImg := os.Remove(imageFile)
		wwlog.Verbose("removing %s for container %s", imageFile+".gz", name)
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

func SetProfileDefaultContainer(name string) error {
	nodeDB, err := node.New()
	if err != nil {
		return fmt.Errorf("failed to create new node, err: %s", err)
	}

	profiles, err := nodeDB.MapAllProfiles()
	if err != nil {
		return fmt.Errorf("failed to map all profiles, err: %s", err)
	}

	defaultProfile := profiles["default"]
	if defaultProfile == nil {
		return fmt.Errorf("failed to get the 'default' profile")
	}

	defaultProfile.ContainerName.Set(name)
	err = nodeDB.ProfileUpdate(*defaultProfile)
	if err != nil {
		return errors.Wrap(err, "failed to update default profile")
	}

	return nodeDB.Persist()
}
