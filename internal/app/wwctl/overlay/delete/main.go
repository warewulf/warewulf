package delete

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/assets"
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
	"path"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var overlayPath string
	config := config.New()

	if SystemOverlay == true {
		overlayPath = config.SystemOverlaySource(args[0])
	} else {
		overlayPath = config.RuntimeOverlaySource(args[0])
	}

	if overlayPath == "" {
		wwlog.Printf(wwlog.ERROR, "Overlay name did not render: '%s'\n", args[0])
		os.Exit(1)
	}

	if util.IsDir(overlayPath) == false {
		wwlog.Printf(wwlog.ERROR, "Overlay name does not exist: '%s'\n", args[0])
		os.Exit(1)
	}

	if len(args) == 1 {
		if Force == true {
			err := os.RemoveAll(overlayPath)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "Failed deleting overlay: %s\n", args[0])
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		} else {
			err := os.Remove(overlayPath)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "Failed deleting overlay: %s\n", args[0])
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		fmt.Printf("Deleted overlay: %s\n", args[0])

	} else if len(args) > 1 {
		for i := 1; i < len(args); i++ {
			removePath := path.Join(overlayPath, args[i])

			if util.IsDir(removePath) == true || util.IsFile(removePath) == true {
				wwlog.Printf(wwlog.ERROR, "Path to remove doesn't exist in overlay: %s\n", removePath)
				os.Exit(1)
			}

			if Force == true {
				err := os.RemoveAll(removePath)
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "Failed deleting file from overlay: %s:%s\n", args[0], args[i])
					wwlog.Printf(wwlog.ERROR, "%s\n", err)
					os.Exit(1)
				}
			} else {
				err := os.Remove(removePath)
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "Failed deleting overlay: %s:%s\n", args[0], args[i])
					wwlog.Printf(wwlog.ERROR, "%s\n", err)
					os.Exit(1)
				}
			}

			if Parents == true {
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
		fmt.Printf("Deleted from overlay: %s:%s\n", args[0], args[1])

	}

	if NoOverlayUpdate == false {
		nodes, err := assets.FindAllNodes()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Cloud not get nodeList: %s\n", err)
			os.Exit(1)
		}

		var updateNodes []assets.NodeInfo

		for _, node := range nodes {
			if SystemOverlay == true && node.SystemOverlay == args[0] {
				updateNodes = append(updateNodes, node)
			} else if node.RuntimeOverlay == args[0] {
				updateNodes = append(updateNodes, node)
			}
		}

		if SystemOverlay == true {
			wwlog.Printf(wwlog.INFO, "Updating System Overlays...\n")
			return overlay.SystemBuild(updateNodes, true)
		} else {
			wwlog.Printf(wwlog.INFO, "Updating Runtime Overlays...\n")
			return overlay.RuntimeBuild(updateNodes, true)
		}
	}

	return nil
}