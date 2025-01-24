package rename

import (
	"fmt"

	"github.com/spf13/cobra"
	api "github.com/warewulf/warewulf/internal/pkg/api/image"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	if len(args) != 2 {
		return fmt.Errorf("rename requires 2 arguments: %d provided", len(args))
	}

	crp := &wwapiv1.ImageRenameParameter{
		ImageName:  args[0],
		TargetName: args[1],
		Build:      SetBuild,
	}

	if !image.DoesSourceExist(crp.ImageName) {
		return fmt.Errorf("%s source dir does not exist", crp.ImageName)
	}

	if image.DoesSourceExist(crp.TargetName) {
		return fmt.Errorf("an other image with the name %s already exists", crp.TargetName)
	}

	if !image.ValidName(crp.TargetName) {
		return fmt.Errorf("image name contains illegal characters : %s", crp.TargetName)
	}

	err = api.ImageRename(crp)
	if err != nil {
		err = fmt.Errorf("could not rename image: %s", err.Error())
		return
	}

	wwlog.Info("Image %s successfully renamed as %s", crp.ImageName, crp.TargetName)
	return
}
