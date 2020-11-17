package edit

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
	"path"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	config := config.New()
	editor := config.Editor
	var overlaySourceDir string


	if len(args) < 2 {
		fmt.Printf("wwctl overlay edit [overlay name] [overlay file]\n")
		cmd.Help()
		os.Exit(1)
	}

	if SystemOverlay == true {
		overlaySourceDir = config.SystemOverlaySource(args[0])
	} else {
		overlaySourceDir = config.RuntimeOverlaySource(args[0])
	}

	if util.IsDir(overlaySourceDir) == false {
		wwlog.Printf(wwlog.ERROR, "Overlay does not exist: %s\n", args[0])
		os.Exit(1)
	}

	overlayFile := path.Join(overlaySourceDir, args[1])

	wwlog.Printf(wwlog.DEBUG, "Will edit overlay file: %s\n", overlayFile)

	err := os.MkdirAll(path.Dir(overlayFile), 0755)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not create directory: %s\n", path.Dir(overlayFile))
		os.Exit(1)
	}

	if editor == "" {
		wwlog.Printf(wwlog.WARN, "No default editor provided, will use `nano`.")
		editor = "nano"
	}

	return util.ExecInteractive(editor, overlayFile)

}

