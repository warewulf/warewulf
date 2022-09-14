package delete

import (
	"fmt"
	"os"
	"path"

	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var overlayPath string
	var fileName string

	overlayName := args[0]

	if len(args) == 2 {
		fileName = args[1]
	}

	overlayPath = overlay.OverlaySourceDir(overlayName)

	if overlayPath == "" {
		wwlog.Error("Overlay name did not resolve: '%s'\n", overlayName)
		os.Exit(1)
	}

	if !util.IsDir(overlayPath) {
		wwlog.Error("Overlay does not exist: %s\n", overlayName)
		os.Exit(1)
	}

	if fileName == "" {
		if overlayName == "wwinit" || overlayName == "host" {
			return errors.New("refusing to delete the Warewulf overlay")
		}
		if Force {
			err := os.RemoveAll(overlayPath)
			if err != nil {
				return errors.Wrap(err, "failed deleting overlay")
			}
		} else {
			err := os.Remove(overlayPath)
			if err != nil {
				return errors.Wrap(err, "failed deleting overlay")
			}
		}
		fmt.Printf("Deleted overlay: %s\n", args[0])

	} else {
		removePath := path.Join(overlayPath, fileName)

		if !util.IsDir(removePath) && !util.IsFile(removePath) {
			wwlog.Error("Path to remove doesn't exist in overlay: %s\n", removePath)
			os.Exit(1)
		}

		if Force {
			err := os.RemoveAll(removePath)
			if err != nil {
				wwlog.Error("Failed deleting file from overlay: %s:%s\n", overlayName, overlayPath)
				wwlog.Error("%s\n", err)
				os.Exit(1)
			}
		} else {
			err := os.Remove(removePath)
			if err != nil {
				wwlog.Error("Failed deleting overlay: %s:%s\n", overlayName, overlayPath)
				wwlog.Error("%s\n", err)
				os.Exit(1)
			}
		}

		if Parents {
			// Cleanup any empty directories left behind...
			i := path.Dir(removePath)
			for i != overlayPath {
				wwlog.Debug("Evaluating directory to remove: %s\n", i)
				err := os.Remove(i)
				if err != nil {
					break
				}

				wwlog.Verbose("Removed empty directory: %s\n", i)
				i = path.Dir(i)
			}
		}
	}

	return nil
}
