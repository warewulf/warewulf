package copy

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	imageSource := args[0]
	imageDestination := args[1]

	if !image.DoesSourceExist(imageSource) {
		return fmt.Errorf("image's source doesn't exists: %s", imageSource)
	}

	if !image.ValidName(imageDestination) {
		return fmt.Errorf("image name contains illegal characters : %s", imageDestination)
	}

	if image.DoesSourceExist(imageDestination) {
		return fmt.Errorf("an other image with name: %s already exists in sources", imageDestination)
	}

	err = image.Duplicate(imageSource, imageDestination)
	if err != nil {
		return fmt.Errorf("could not duplicate image: %s", err.Error())
	}

	if Build {
		err = image.Build(imageDestination, true)
		if err != nil {
			return err
		}
	}

	wwlog.Info("Image %s successfully duplicated as %s", imageSource, imageDestination)
	return
}
