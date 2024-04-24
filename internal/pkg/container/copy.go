package container

import (
	"fmt"

	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

type CopyParameter struct {
	Name        string
	Destination string
	ForceBuild  bool
}

func Copy(param *CopyParameter) error {
	if !DoesSourceExist(param.Name) {
		return fmt.Errorf("copy source %s does not exists", param.Name)
	}

	if !ValidName(param.Destination) {
		return fmt.Errorf("copy destination %s contains illegal chracters", param.Destination)
	}

	if DoesContainerExists(param.Destination) {
		return fmt.Errorf("destination %s already exists", param.Destination)
	}

	return duplicate(param.Name, param.Destination, param.ForceBuild)
}

func duplicate(name string, destination string, build bool) error {
	fullPathImageSource := RootFsDir(name)

	wwlog.Info("Copying sources...")
	err := ImportDirectory(fullPathImageSource, destination)
	if err != nil {
		return err
	}

	if build {
		wwlog.Info("Building container: %s", destination)
		return Build(&BuildParameter{
			Names: []string{destination},
			Force: true,
		})
	}

	return nil
}
