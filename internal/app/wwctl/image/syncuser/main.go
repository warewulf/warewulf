package syncuser

import (
	"fmt"

	"github.com/spf13/cobra"
	image_build "github.com/warewulf/warewulf/internal/pkg/api/image"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	imageName := args[0]
	if !image.ValidName(imageName) {
		return fmt.Errorf("%s is not a valid image", imageName)
	}
	err := image.Syncuser(imageName, write)
	if err != nil {
		return fmt.Errorf("error in synchronize: %s", err)
	}

	if write && !build {
		// when write = true and build = false, we will print a warnning, this is the default case
		wwlog.Warn("Syncuser is completed. Rebuild image or add `--build` flag for automatic rebuild after syncuser.")
	} else if write && build {
		// if write = true and build = true, then it'll trigger the image build after sync
		cbp := &wwapiv1.ImageBuildParameter{
			ImageNames: []string{imageName},
			Force:      true,
			All:        false,
		}
		err := image_build.ImageBuild(cbp)
		if err != nil {
			return fmt.Errorf("error during image build: %s", err)
		}
	}

	return nil
}
