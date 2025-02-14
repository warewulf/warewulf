package build

import (
	"fmt"

	"github.com/spf13/cobra"
	apiimage "github.com/warewulf/warewulf/internal/pkg/api/image"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/warewulf/warewulf/internal/pkg/image"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	if SyncUser {
		for _, name := range args {
			if err := image.Syncuser(name, true); err != nil {
				return fmt.Errorf("syncuser error: %w", err)
			}
		}
	}

	cbp := &wwapiv1.ImageBuildParameter{
		ImageNames: args,
		Force:      BuildForce,
		All:        BuildAll,
	}
	return apiimage.ImageBuild(cbp)
}
