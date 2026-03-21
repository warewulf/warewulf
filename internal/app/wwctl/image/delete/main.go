package delete

import (
	"fmt"

	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/util"

	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, imageNames []string) (err error) {
	if !SetYes {
		yes := util.Confirm(fmt.Sprintf("Are you sure you want to delete image %s", imageNames))
		if !yes {
			return
		}

	}

	for _, imageName := range imageNames {
		if err := image.Delete(imageName); err != nil {
			return fmt.Errorf("error deleting image %s: %s", imageName, err)
		}
	}

	return nil
}
