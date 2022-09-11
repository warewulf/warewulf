package edit

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	editor := os.Getenv("EDITOR")
	var overlaySourceDir string

	overlayName := args[0]
	fileName := args[1]

	if editor == "" {
		editor = "/bin/vi"
	}

	overlaySourceDir = overlay.OverlaySourceDir(overlayName)

	if !util.IsDir(overlaySourceDir) {
		wwlog.Error("Overlay does not exist: %s\n", overlayName)
		os.Exit(1)
	}

	overlayFile := path.Join(overlaySourceDir, fileName)

	wwlog.Debug("Will edit overlay file: %s\n", overlayFile)

	if CreateDirs {
		err := os.MkdirAll(path.Dir(overlayFile), 0755)
		if err != nil {
			wwlog.Error("Could not create directory: %s\n", path.Dir(overlayFile))
			os.Exit(1)
		}
	} else {
		if !util.IsDir(path.Dir(overlayFile)) {
			wwlog.Error("Can not create file, parent directory does not exist, try adding the\n")
			wwlog.Error("'--parents' option to create the directory.\n")
			os.Exit(1)
		}
	}

	if !util.IsFile(overlayFile) && filepath.Ext(overlayFile) == ".ww" {
		wwlog.Verbose("This is a new file, creating some default content\n")

		w, err := os.OpenFile(overlayFile, os.O_RDWR|os.O_CREATE, os.FileMode(PermMode))
		if err != nil {
			wwlog.Warn("Could not create file for writing: %s\n", err)
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

	err := util.ExecInteractive(editor, overlayFile)
	if err != nil {
		wwlog.Error("Editor process existed with non-zero\n")
		os.Exit(1)
	}

	return nil
}
