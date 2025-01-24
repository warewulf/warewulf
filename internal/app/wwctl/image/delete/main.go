package delete

import (
	"fmt"

	"github.com/warewulf/warewulf/internal/pkg/api/image"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/warewulf/warewulf/internal/pkg/util"

	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	cdp := &wwapiv1.ImageDeleteParameter{
		ImageNames: args,
	}

	if !SetYes {
		yes := util.Confirm(fmt.Sprintf("Are you sure you want to delete image %s", args))
		if !yes {
			return
		}

	}
	return image.ImageDelete(cdp)
}
