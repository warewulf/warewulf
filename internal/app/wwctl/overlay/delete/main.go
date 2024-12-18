package delete

import (
	"fmt"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var overlayPath string
	var fileName string

	overlayName := args[0]

	if len(args) == 2 {
		fileName = args[1]
	}

	overlayPath, isSite := overlay.GetOverlay(overlayName)
	if !isSite {
		return fmt.Errorf("distribution overlay can't deleted")

	}
	if overlayPath == "" {
		return fmt.Errorf("overlay name did not resolve: '%s'", overlayName)
	}

	if !util.IsDir(overlayPath) {
		return fmt.Errorf("overlay does not exist: %s", overlayName)
	}

	if fileName == "" {
		if overlayName == "wwinit" || overlayName == "host" {
			return errors.New("refusing to delete the Warewulf overlay")
		}
		if Force {
			err := os.RemoveAll(overlayPath)
			if err != nil {
				return fmt.Errorf("failed deleting overlay: %w", err)
			}
		} else {
			err := os.Remove(overlayPath)
			if err != nil {
				return fmt.Errorf("failed deleting overlay: %w", err)
			}
		}
		wwlog.Info("Deleted overlay: %s\n", args[0])

	} else {
		removePath := path.Join(overlayPath, fileName)

		if !util.IsDir(removePath) && !util.IsFile(removePath) {
			return fmt.Errorf("path to remove doesn't exist in overlay: %s", removePath)
		}

		if Force {
			err := os.RemoveAll(removePath)
			if err != nil {
				return fmt.Errorf("failed deleting file from overlay: %s:%s", overlayName, overlayPath)
			}
		} else {
			err := os.Remove(removePath)
			if err != nil {
				return fmt.Errorf("failed deleting overlay: %s:%s", overlayName, overlayPath)
			}
		}

		if Parents {
			// Cleanup any empty directories left behind...
			i := path.Dir(removePath)
			for i != overlayPath {
				wwlog.Debug("Evaluating directory to remove: %s", i)
				err := os.Remove(i)
				if err != nil {
					break
				}

				wwlog.Verbose("Removed empty directory: %s", i)
				i = path.Dir(i)
			}
		}
	}

	return nil
}
