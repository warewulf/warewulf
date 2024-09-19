package copy

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/warewulf/warewulf/internal/pkg/container"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {

	if len(args) > 2 {
		wwlog.Warn("copy only requires 2 arguments but you provided %d arguments. Hence, they will be ignored.", len(args))
	}

	cdp := &wwapiv1.ContainerCopyParameter{
		ContainerSource:      args[0],
		ContainerDestination: args[1],
	}

	if !container.DoesSourceExist(cdp.ContainerSource) {
		return fmt.Errorf("container's source doesn't exists: %s", cdp.ContainerSource)
	}

	if !container.ValidName(cdp.ContainerDestination) {
		return fmt.Errorf("container name contains illegal characters : %s", cdp.ContainerDestination)
	}

	if container.DoesSourceExist(cdp.ContainerDestination) {
		return fmt.Errorf("an other container with name: %s already exists in sources", cdp.ContainerDestination)
	}

	err = container.Duplicate(cdp.ContainerSource, cdp.ContainerDestination)
	if err != nil {
		return fmt.Errorf("could not duplicate image: %s", err.Error())
	}

	wwlog.Info("Container %s successfully duplicated as %s", cdp.ContainerSource, cdp.ContainerDestination)
	return
}
