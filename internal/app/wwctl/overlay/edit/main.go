package edit

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
	"path/filepath"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	config := config.New()
	editor := config.Editor
	var overlaySourceDir string


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

	if CreateDirs == true {
		err := os.MkdirAll(path.Dir(overlayFile), 0755)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not create directory: %s\n", path.Dir(overlayFile))
			os.Exit(1)
		}
	} else {
		if util.IsDir(path.Dir(overlayFile)) == false {
			wwlog.Printf(wwlog.ERROR, "Can not create file, parent directory does not exist, try adding the\n")
			wwlog.Printf(wwlog.ERROR, "'--parents' option to create the directory.\n")
			os.Exit(1)
		}
	}

	if util.IsFile(overlayFile) == false && filepath.Ext(overlayFile) == ".ww" {
		wwlog.Printf(wwlog.WARN, "This is a new file, creating some default content\n")

		w, err := os.OpenFile(overlayFile, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			wwlog.Printf(wwlog.WARN, "Could not create file for writing: %s\n", err)
		}

		fmt.Fprintf(w, "# This is a Warewulf Template file.\n")
		fmt.Fprintf(w, "#\n")
		fmt.Fprintf(w, "# This file (suffix '.ww') will be automatically rewritten without the suffix\n")
		fmt.Fprintf(w, "# when the overlay is rendered for the individual nodes. Here are some examples\n")
		fmt.Fprintf(w, "# of macros and logic which can be used within this file:\n")
		fmt.Fprintf(w, "#\n")
		fmt.Fprintf(w, "# Node FQDN = {{.Fqdn}}\n")
		fmt.Fprintf(w, "# Node Group = {{.GroupName}}\n")
		fmt.Fprintf(w, "# Network Config = {{.NetDevs.eth0.Ipaddr}}, {{.NetDevs.eth0.Hwaddr}}, etc.\n")
		fmt.Fprintf(w, "#\n")
		fmt.Fprintf(w, "# Goto the documentation pages for more information: http://www.hpcng.org/...\n")
		fmt.Fprintf(w, "\n")
	}

	err := util.ExecInteractive(editor, overlayFile)

	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Editor process existed with non-zero\n")
		os.Exit(1)
	}

	// Everything below this point is to update the relevant overlays
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

	return nil
}

