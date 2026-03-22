package image

import (
	"fmt"
	"os"

	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Delete(name string) error {
	// validate image names
	if !ValidSource(name) {
		return fmt.Errorf("image name is not valid source: %s", name)
	}

	nodeDB, err := node.New()
	if err != nil {
		return fmt.Errorf("could not open nodeDB: %s", err)
	}

	// check if the deleted images are not used by nodes
	for nodeName, node := range nodeDB.Nodes {
		if node.ImageName == name {
			return fmt.Errorf("image %s is in use by node %s, cannot delete", node.ImageName, nodeName)
		}
	}

	// check if the deleted images are not used by profiles
	for profileName, profile := range nodeDB.NodeProfiles {
		if profile.ImageName == name {
			return fmt.Errorf("image %s is in use by profile %s, cannot delete", profile.ImageName, profileName)
		}
	}

	// delete images
	if err := DeleteSource(name); err != nil {
		return fmt.Errorf("could not remove image source %s: %w", name, err)
	}
	if err := DeleteImage(name); err != nil {
		return fmt.Errorf("could not remove image file %s: %w", name, err)
	}
	wwlog.Info("Deleted image %q", name)

	return nil
}

/*
Delete the chroot of an image
*/
func DeleteSource(name string) error {
	fullPath := SourceDir(name)

	wwlog.Verbose("Removing path: %s", fullPath)
	return os.RemoveAll(fullPath)
}

/*
Delete the image of an image
*/
func DeleteImage(name string) error {
	imageFile := ImageFile(name)
	if util.IsFile(imageFile) {
		wwlog.Verbose("removing %s for image %s", imageFile, name)
		errImg := os.Remove(imageFile)
		wwlog.Verbose("removing %s for image %s", imageFile+".gz", name)
		errGz := os.Remove(imageFile + ".gz")
		if errImg != nil {
			return fmt.Errorf("problem deleting %s for image %s: %s", imageFile, name, errImg)
		}
		if errGz != nil {
			return fmt.Errorf("problem deleting %s for image %s: %s", imageFile+".gz", name, errGz)
		}
		return nil
	}
	return fmt.Errorf("image %s of image %s doesn't exist", imageFile, name)
}
