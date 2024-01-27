package build

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/api/container"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	cbp := &wwapiv1.ContainerBuildParameter{
		ContainerNames: args,
		Force:          BuildForce,
		All:            BuildAll,
		Default:        SetDefault,
	}
	return container.ContainerBuild(cbp)
}
