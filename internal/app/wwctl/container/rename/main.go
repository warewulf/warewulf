package rename

import (
	"fmt"

	api "github.com/hpcng/warewulf/internal/pkg/api/container"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	if len(args) > 2 {
		wwlog.Warn("rename only requires 2 arguments but you provided %d arguments. Hence, they will be ignored.", len(args))
	}

	crp := &wwapiv1.ContainerRenameParameter{
		ContainerName: args[0],
		TargetName:    args[1],
		Build:         SetBuild,
	}

	if !container.DoesSourceExist(crp.ContainerName) {
		return fmt.Errorf("%s source dir does not exist", crp.ContainerName)
	}

	if container.DoesSourceExist(crp.TargetName) {
		return fmt.Errorf("an other container with the name %s already exists", crp.TargetName)
	}

	if !container.ValidName(crp.TargetName) {
		wwlog.Error("Container name contains illegal characters : %s", crp.TargetName)
		return
	}

	err = api.ContainerRename(crp)
	if err != nil {
		err = fmt.Errorf("could not rename image: %s", err.Error())
		return
	}

	wwlog.Info("Container %s successfully renamed as %s", crp.ContainerName, crp.TargetName)
	return
}
