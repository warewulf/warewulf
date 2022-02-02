package show

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var overlaySourceDir string

	overlayName := args[0]
	fileName := args[1]

	overlaySourceDir = overlay.OverlaySourceDir(overlayName)

	if !util.IsDir(overlaySourceDir) {
		wwlog.Printf(wwlog.ERROR, "Overlay does not exist: %s\n", overlayName)
		os.Exit(1)
	}

	overlayFile := path.Join(overlaySourceDir, fileName)

	if !util.IsFile(overlayFile) {
		wwlog.Printf(wwlog.ERROR, "File does not exist within overlay: %s:%s\n", overlayName, fileName)
		os.Exit(1)
	}

	f, err := ioutil.ReadFile(overlayFile)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not read file: %s\n", err)
		os.Exit(1)
	}

	fmt.Print(string(f))

	return nil
}
