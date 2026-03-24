package image

import (
	"fmt"
	"os"

	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Rename(name string, targetName string, build bool) error {
	if !ValidSource(name) {
		return fmt.Errorf("image source does not exist: %s", name)
	}
	if !ValidName(targetName) {
		return fmt.Errorf("invalid image name: %s", targetName)
	}

	// rename the image source folder
	sourceDir := SourceDir(name)
	destDir := SourceDir(targetName)
	err := os.Rename(sourceDir, destDir)
	if err != nil {
		return err
	}

	err = DeleteImage(name)
	if err != nil {
		wwlog.Warn("Could not remove image files for %s: %s", name, err)
	}

	if build {
		err = Build(targetName, true)
		if err != nil {
			return err
		}
	}

	// update the nodes profiles image name
	nodeDB, err := node.New()
	if err != nil {
		return err
	}

	for nodeId, node := range nodeDB.Nodes {
		if node.ImageName == name {
			wwlog.Debug("updating node %s image to %s", nodeId, targetName)
			nodeDB.Nodes[nodeId].ImageName = targetName
		}
	}

	for profileId, profile := range nodeDB.NodeProfiles {
		if profile.ImageName == name {
			wwlog.Debug("updating profile %s image to %s", profileId, targetName)
			nodeDB.NodeProfiles[profileId].ImageName = targetName
		}
	}

	return nodeDB.Persist()
}
