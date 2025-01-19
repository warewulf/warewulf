package delete

import (
	"fmt"

	"github.com/warewulf/warewulf/internal/pkg/api/image"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	apiutil "github.com/warewulf/warewulf/internal/pkg/api/util"

	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	cdp := &wwapiv1.ImageDeleteParameter{
		ImageNames: args,
	}

	if !SetYes {
		yes := apiutil.ConfirmationPrompt(fmt.Sprintf("Are you sure you want to delete image %s", args))
		if !yes {
			return
		}

	}
	return image.ImageDelete(cdp)
}
