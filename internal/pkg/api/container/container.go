package container

import (
	"fmt"

	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/warewulf/warewulf/internal/pkg/container"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func ContainerCopy(cbp *wwapiv1.ContainerCopyParameter) (err error) {
	if cbp == nil {
		return fmt.Errorf("ContainerCopyParameter is nil")
	}

	err = container.Copy(&container.CopyParameter{
		Name:        cbp.ContainerSource,
		Destination: cbp.ContainerDestination,
		ForceBuild:  true,
	})
	if err != nil {
		err = fmt.Errorf("could not copy image: %s", err.Error())
		wwlog.Error(err.Error())
		return
	}

	wwlog.Info("Container %s successfully duplicated as %s", cbp.ContainerSource, cbp.ContainerDestination)
	return
}

func ContainerBuild(cbp *wwapiv1.ContainerBuildParameter) (err error) {
	if cbp == nil {
		return fmt.Errorf("ContainerBuildParameter is nil")
	}

	var containers []string

	if cbp.All {
		containers, err = container.ListSources()
		if err != nil {
			return
		}
	} else {
		containers = cbp.ContainerNames
	}

	return container.Build(&container.BuildParameter{
		Names:   containers,
		Force:   cbp.Force,
		Default: cbp.Default,
	})
}

func ContainerDelete(cdp *wwapiv1.ContainerDeleteParameter) (err error) {
	if cdp == nil {
		return fmt.Errorf("ContainerDeleteParameter is nil")
	}

	return container.Delete(&container.DeleteParameter{
		Names: cdp.ContainerNames,
	})
}

func ContainerImport(cip *wwapiv1.ContainerImportParameter) error {
	if cip == nil {
		return fmt.Errorf("input param is nil")
	}

	err := container.Import(&container.ImportParameter{
		Name:     cip.Name,
		Source:   cip.Source,
		Force:    cip.Force,
		Update:   cip.Update,
		Build:    cip.Build,
		Default:  cip.Default,
		SyncUser: cip.SyncUser,
	})

	if err != nil {
		return err
	}

	// we need to reload the daemon to reflect profile container changes
	if cip.Default {
		err = warewulfd.DaemonStatus()
		if err != nil {
			// warewulfd is not running, skip
			return nil
		}
		return warewulfd.DaemonReload()
	}

	return nil
}

func ContainerList() (containerInfo []*wwapiv1.ContainerInfo, err error) {
	containers, err := container.List()
	if err != nil {
		return nil, err
	}
	for _, container := range containers {
		containerInfo = append(containerInfo, &wwapiv1.ContainerInfo{
			Name:          container.Name,
			NodeCount:     container.NodeCount,
			KernelVersion: container.KernelVersion,
			CreateDate:    container.CreateDate,
			ModDate:       container.ModDate,
			Size:          container.Size,
		})
	}
	return
}

func ContainerShow(csp *wwapiv1.ContainerShowParameter) (response *wwapiv1.ContainerShowResponse, err error) {
	resp, err := container.Show(&container.ShowParameter{
		Name: csp.ContainerName,
	})

	if err != nil {
		return nil, err
	}

	return &wwapiv1.ContainerShowResponse{
		Name:          resp.Name,
		Rootfs:        resp.Rootfs,
		Nodes:         resp.Nodes,
		KernelVersion: resp.KernelVersion,
	}, nil
}

func ContainerRename(crp *wwapiv1.ContainerRenameParameter) (err error) {
	err = container.Rename(&container.RenameParameter{
		Name:       crp.ContainerName,
		TargetName: crp.TargetName,
		Build:      crp.Build,
	})

	if err != nil {
		return err
	}

	err = warewulfd.DaemonStatus()
	if err != nil {
		// warewulfd is not running, skip
		return nil
	}

	// else reload daemon to apply new changes
	return warewulfd.DaemonReload()
}
