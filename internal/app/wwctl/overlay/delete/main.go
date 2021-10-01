package delete

import (
	"fmt"
	"os"
	"path"

	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var overlayPath string
	var fileName string

	overlayKind := args[0]
	overlayName := args[1]

	if len(args) == 3 {
		fileName = args[2]
	}

	if overlayName == "default" {
		return errors.New("refusing to delete the default overlay")
	}

	if overlayKind != "system" && overlayKind != "runtime" {
		return errors.New("overlay kind must be of type 'system' or 'runtime'")
	}

	if overlayKind == "system" {
		overlayPath = config.SystemOverlaySource(overlayName)
	} else if overlayKind == "runtime" {
		overlayPath = config.RuntimeOverlaySource(overlayName)
	}

	if overlayPath == "" {
		wwlog.Printf(wwlog.ERROR, "Overlay name did not resolve: '%s'\n", overlayName)
		os.Exit(1)
	}

	if !util.IsDir(overlayPath) {
		wwlog.Printf(wwlog.ERROR, "Overlay does not exist: '%s:%s'\n", overlayKind, overlayName)
		os.Exit(1)
	}

	if fileName == "" {
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

		n, err := node.New()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
			os.Exit(1)
		}

		nodes, err := n.FindAllNodes()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not get node list: %s\n", err)
			os.Exit(1)
		}

		for _, node := range nodes {
			if overlayKind == "system" && node.SystemOverlay.Get() == overlayName {
				node.SystemOverlay.Set("default")
			} else if overlayKind == "runtime" && node.RuntimeOverlay.Get() == overlayName {
				node.RuntimeOverlay.Set("default")
			}
		}

		err = n.Persist()
		if err != nil {
			return errors.Wrap(err, "failed to persist node updates")
		}

	} else {
		removePath := path.Join(overlayPath, fileName)

		if !util.IsDir(removePath) && !util.IsFile(removePath) {
			wwlog.Printf(wwlog.ERROR, "Path to remove doesn't exist in overlay: %s\n", removePath)
			os.Exit(1)
		}

		if Force {
			err := os.RemoveAll(removePath)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "Failed deleting file from overlay: %s:%s:%s\n", overlayKind, overlayName, overlayPath)
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		} else {
			err := os.Remove(removePath)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "Failed deleting overlay: %s:%s:%s\n", overlayKind, overlayName, overlayPath)
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}

		if Parents {
			// Cleanup any empty directories left behind...
			i := path.Dir(removePath)
			for i != overlayPath {
				wwlog.Printf(wwlog.DEBUG, "Evaluating directory to remove: %s\n", i)
				err := os.Remove(i)
				if err != nil {
					break
				}

				wwlog.Printf(wwlog.VERBOSE, "Removed empty directory: %s\n", i)
				i = path.Dir(i)
			}
		}
	}

	return nil
}
