package container

import (
	"fmt"
	"os"

	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

type RenameParameter struct {
	Name       string
	TargetName string
	Build      bool
}

func Rename(param *RenameParameter) error {
	if !DoesSourceExist(param.Name) {
		return fmt.Errorf("container name %s does not exist", param.Name)
	}

	if DoesSourceExist(param.TargetName) {
		return fmt.Errorf("another container with the name %s already exists", param.TargetName)
	}

	if !ValidName(param.TargetName) {
		return fmt.Errorf("container name contains illegal characters : %s", param.TargetName)
	}

	// rename the container source folder
	sourceDir := SourceDir(param.Name)
	destDir := SourceDir(param.TargetName)
	err := os.Rename(sourceDir, destDir)
	if err != nil {
		return err
	}

	err = DeleteImage(param.Name)
	if err != nil {
		wwlog.Warn("Could not remove image files for %s: %w", param.Name, err)
	}

	if param.Build {
		err = Build(&BuildParameter{
			Names: []string{param.TargetName},
			Force: true,
		})
		if err != nil {
			return err
		}
	}

	// update the nodes profiles container name
	nodeDB, err := node.New()
	if err != nil {
		return err
	}

	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		return err
	}
	for _, node := range nodes {
		if node.ContainerName.Get() == param.Name {
			node.ContainerName.Set(param.TargetName)
			if err := nodeDB.NodeUpdate(node); err != nil {
				return err
			}
		}
	}

	profiles, err := nodeDB.FindAllProfiles()
	if err != nil {
		return err
	}
	for _, profile := range profiles {
		if profile.ContainerName.Get() == param.Name {
			profile.ContainerName.Set(param.TargetName)
			if err := nodeDB.ProfileUpdate(profile); err != nil {
				return err
			}
		}
	}

	return nodeDB.Persist()
}
