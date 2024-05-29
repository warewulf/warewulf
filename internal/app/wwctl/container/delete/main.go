package delete

import (
	"fmt"

	apiutil "github.com/warewulf/warewulf/internal/pkg/api/util"
	"github.com/warewulf/warewulf/internal/pkg/container"

	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	if !SetYes {
		yes := apiutil.ConfirmationPrompt(fmt.Sprintf("Are you sure you want to delete container %s", args))
		if !yes {
			return
		}

	}
	return container.Delete(&container.DeleteParameter{
		Names: args,
	})
}
