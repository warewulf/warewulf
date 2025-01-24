package build

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/api/image"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	cbp := &wwapiv1.ImageBuildParameter{
		ImageNames: args,
		Force:      BuildForce,
		All:        BuildAll,
	}
	return image.ImageBuild(cbp)
}
