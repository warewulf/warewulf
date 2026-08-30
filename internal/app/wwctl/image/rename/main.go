package rename

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	if !image.DoesSourceExist(args[0]) {
		return fmt.Errorf("%s source dir does not exist", args[0])
	}

	if image.DoesSourceExist(args[1]) {
		return fmt.Errorf("an other image with the name %s already exists", args[1])
	}

	if !image.ValidName(args[1]) {
		return fmt.Errorf("image name contains illegal characters: %s", args[1])
	}

	err = image.Rename(args[0], args[1], SetBuild)
	if err != nil {
		err = fmt.Errorf("could not rename image: %s", err.Error())
		return
	}

	wwlog.Info("Image %s successfully renamed as %s", args[0], args[1])

	if err = warewulfd.DaemonStatus(); err != nil {
		return nil
	}
	return warewulfd.DaemonReload()
}
