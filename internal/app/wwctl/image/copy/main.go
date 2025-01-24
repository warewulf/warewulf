package copy

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {

	if len(args) > 2 {
		wwlog.Warn("copy only requires 2 arguments but you provided %d arguments. Hence, they will be ignored.", len(args))
	}

	cdp := &wwapiv1.ImageCopyParameter{
		ImageSource:      args[0],
		ImageDestination: args[1],
		Build:            Build,
	}

	if !image.DoesSourceExist(cdp.ImageSource) {
		return fmt.Errorf("image's source doesn't exists: %s", cdp.ImageSource)
	}

	if !image.ValidName(cdp.ImageDestination) {
		return fmt.Errorf("image name contains illegal characters : %s", cdp.ImageDestination)
	}

	if image.DoesSourceExist(cdp.ImageDestination) {
		return fmt.Errorf("an other image with name: %s already exists in sources", cdp.ImageDestination)
	}

	err = image.Duplicate(cdp.ImageSource, cdp.ImageDestination)
	if err != nil {
		return fmt.Errorf("could not duplicate image: %s", err.Error())
	}

	if cdp.Build {
		err = image.Build(cdp.ImageDestination, true)
		if err != nil {
			return err
		}
	}

	wwlog.Info("Image %s successfully duplicated as %s", cdp.ImageSource, cdp.ImageDestination)
	return
}
