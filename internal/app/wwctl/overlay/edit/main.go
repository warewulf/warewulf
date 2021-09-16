package edit

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	editor := os.Getenv("EDITOR")
	var overlaySourceDir string

	overlayKind := args[0]
	overlayName := args[1]
	fileName := args[2]

	if overlayKind != "system" && overlayKind != "runtime" {
		return errors.New("overlay kind must be of type 'system' or 'runtime'")
	}

	if editor == "" {
		editor = "/bin/vi"
	}

	if overlayKind == "system" {
		overlaySourceDir = config.SystemOverlaySource(overlayName)
	} else if overlayKind == "runtime" {
		overlaySourceDir = config.RuntimeOverlaySource(overlayName)
	}

	if !util.IsDir(overlaySourceDir) {
		wwlog.Printf(wwlog.ERROR, "Overlay does not exist: %s:%s\n", overlayKind, overlayName)
		os.Exit(1)
	}

	overlayFile := path.Join(overlaySourceDir, fileName)

	wwlog.Printf(wwlog.DEBUG, "Will edit overlay file: %s\n", overlayFile)

	if CreateDirs {
		err := os.MkdirAll(path.Dir(overlayFile), 0755)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not create directory: %s\n", path.Dir(overlayFile))
			os.Exit(1)
		}
	} else {
		if !util.IsDir(path.Dir(overlayFile)) {
			wwlog.Printf(wwlog.ERROR, "Can not create file, parent directory does not exist, try adding the\n")
			wwlog.Printf(wwlog.ERROR, "'--parents' option to create the directory.\n")
			os.Exit(1)
		}
	}

	if !util.IsFile(overlayFile) && filepath.Ext(overlayFile) == ".ww" {
		wwlog.Printf(wwlog.WARN, "This is a new file, creating some default content\n")

		w, err := os.OpenFile(overlayFile, os.O_RDWR|os.O_CREATE, os.FileMode(PermMode))
		if err != nil {
			wwlog.Printf(wwlog.WARN, "Could not create file for writing: %s\n", err)
		}

		fmt.Fprintf(w, "# This is a Warewulf Template file.\n")
		fmt.Fprintf(w, "#\n")
		fmt.Fprintf(w, "# This file (suffix '.ww') will be automatically rewritten without the suffix\n")
		fmt.Fprintf(w, "# when the overlay is rendered for the individual nodes. Here are some examples\n")
		fmt.Fprintf(w, "# of macros and logic which can be used within this file:\n")
		fmt.Fprintf(w, "#\n")
		fmt.Fprintf(w, "# Node FQDN = {{.Id}}\n")
		fmt.Fprintf(w, "# Node Cluster = {{.ClusterName}}\n")
		fmt.Fprintf(w, "# Network Config = {{.NetDevs.eth0.Ipaddr}}, {{.NetDevs.eth0.Hwaddr}}, etc.\n")
		fmt.Fprintf(w, "#\n")
		fmt.Fprintf(w, "# Goto the documentation pages for more information: http://www.hpcng.org/...\n")
		fmt.Fprintf(w, "\n")
	}

	shasum1, err := util.ShaSumFile(overlayFile)
	if err != nil {
		wwlog.Printf(wwlog.WARN, "Could not open overlay file for checksum: %s\n", err)
	}

	err = util.ExecInteractive(editor, overlayFile)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Editor process existed with non-zero\n")
		os.Exit(1)
	}
	wwlog.Printf(wwlog.INFO, "Updated: %s:%s:%s\n", overlayKind, overlayName, fileName)

	shasum2, err := util.ShaSumFile(overlayFile)
	if err != nil {
		wwlog.Printf(wwlog.WARN, "Could not open overlay file for checksum: %s\n", err)
	}

	if shasum1 == shasum2 {
		wwlog.Printf(wwlog.VERBOSE, "Not updating overlays, no file change\n")
		NoOverlayUpdate = true
	}

	if !NoOverlayUpdate {
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

		var updateNodes []node.NodeInfo

		for _, node := range nodes {
			if overlayKind == "system" && node.SystemOverlay.Get() == overlayName {
				updateNodes = append(updateNodes, node)
			} else if overlayKind == "runtime" && node.RuntimeOverlay.Get() == overlayName {
				updateNodes = append(updateNodes, node)
			}
		}

		if overlayKind == "system" {
			wwlog.Printf(wwlog.INFO, "Updating System Overlays...\n")
			return overlay.BuildSystemOverlay(updateNodes)
		} else if overlayKind == "runtime" {
			wwlog.Printf(wwlog.INFO, "Updating Runtime Overlays...\n")
			return overlay.BuildRuntimeOverlay(updateNodes)
		}
	}

	return nil
}
