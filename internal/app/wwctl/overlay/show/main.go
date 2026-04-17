package show

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	overlayName := args[0]
	fileName := args[1]

	overlay_, err := overlay.Get(overlayName)
	if err != nil {
		return err
	}

	overlayFile := overlay_.File(fileName)

	if NodeName == "" {
		// No node specified: show the raw template source without rendering.
		if !util.IsFile(overlayFile) {
			return fmt.Errorf("%s: %s not found", overlayName, overlayFile)
		}
		f, err := os.ReadFile(overlayFile)
		if err != nil {
			return err
		}

		wwlog.Output("%s", string(f))
	} else {
		// Node specified: render the template for the given node and show output.

		// If the caller gave the output filename (without .ww), find the template.
		if !util.IsFile(overlayFile) {
			possibleFile := fmt.Sprintf("%s.ww", overlayFile)
			if filepath.Ext(overlayFile) != ".ww" && util.IsFile(possibleFile) {
				wwlog.Debug("found overlay template: %s", possibleFile)
				overlayFile = possibleFile
			} else {
				return fmt.Errorf("%s: %s not found", overlayName, overlayFile)
			}
		}
		if filepath.Ext(overlayFile) != ".ww" {
			wwlog.Warn("%s lacks the '.ww' suffix, will not be rendered in an overlay", fileName)
		}
		nodeDB, err := node.New()
		if err != nil {
			return err
		}
		nodeConf, err := nodeDB.GetNode(NodeName)
		if err == node.ErrNotFound {
			// Unknown node name: fall back to the local hostname so operators can
			// preview templates on the warewulf server itself.
			hostName, err := os.Hostname()
			if err != nil {
				return fmt.Errorf("could not get host name: %s", err)
			}
			nodeConf = node.NewNode(hostName)
			nodeConf.ClusterName = hostName
		}
		var allNodes []node.Node
		allNodes, err = nodeDB.FindAllNodes()
		if err != nil {
			return err
		}
		tstruct, err := overlay.InitStruct(overlayName, nodeConf, allNodes)
		if err != nil {
			return err
		}
		tstruct.BuildSource = overlayFile
		rendered, err := overlay.RenderTemplate(overlayFile, tstruct)
		if err != nil {
			return err
		}

		if !rendered.WriteFile {
			// abort() was called in the template: nothing would be written to disk.
			// Still show any content rendered before the abort() call for debugging.
			if !Quiet {
				wwlog.Info("backupFile: %v", rendered.BackupFile)
				wwlog.Info("writeFile: %v", rendered.WriteFile)
			}
			wwlog.Output("%s", rendered.Files[0].Buffer.String())
			if !Quiet {
				wwlog.Info("Aborted")
			}
			return nil
		}

		for _, f := range rendered.Files {
			// The default slot (Name == "") holds pre-file() content or a symlink.
			// When named files are also present, non-symlink content is discarded
			// for disk output; skip it here too so that show output matches what
			// is actually written to disk. A default symlink is always shown.
			if f.Name == "" && len(rendered.Files) > 1 && !f.IsSymlink {
				continue
			}
			// The default slot's display name is the template filename with .ww stripped.
			displayName := f.Name
			if displayName == "" {
				displayName = strings.TrimSuffix(fileName, ".ww")
			}
			if !Quiet {
				wwlog.Info("backupFile: %v", rendered.BackupFile)
				wwlog.Info("writeFile: %v", rendered.WriteFile)
				wwlog.Info("Filename: %s", displayName)
				if f.IsSymlink {
					wwlog.Info("Symlink: %s", f.Target)
				}
			}
			if !f.IsSymlink {
				wwlog.Output("%s", f.Buffer.String())
			}
		}
	}
	return nil
}
