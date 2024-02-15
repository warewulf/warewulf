package syncuser

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	container_build "github.com/warewulf/warewulf/internal/pkg/api/container"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/warewulf/warewulf/internal/pkg/container"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	containerName := args[0]
	if !container.ValidName(containerName) {
		return fmt.Errorf("%s is not a valid container", containerName)
	}
	err := container.SyncUids(containerName, !write)
	if err != nil {
		wwlog.Error("Error in synchronize: %s", err)
		os.Exit(1)
	}

	if write && !build {
		// when write = true and build = false, we will print a warnning, this is the default case
		wwlog.Warn("Syncuser is completed, please remember to rebuild container or add `--build` flag for automatic rebuild after syncuser")
	} else if write && build {
		// if write = true and build = true, then it'll trigger the container build after sync
		cbp := &wwapiv1.ContainerBuildParameter{
			ContainerNames: []string{containerName},
			Force:          true,
			All:            false,
			Default:        false,
		}
		err := container_build.ContainerBuild(cbp)
		if err != nil {
			wwlog.Error("Error during container build: %s", err)
			os.Exit(1)
		}
	}

	return nil
}
