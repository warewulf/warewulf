package show

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/node"
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
		wwlog.Error("Overlay does not exist: %s", overlayName)
		os.Exit(1)
	}

	overlayFile := path.Join(overlaySourceDir, fileName)

	if !util.IsFile(overlayFile) {
		wwlog.Error("File does not exist within overlay: %s:%s", overlayName, fileName)
		os.Exit(1)
	}

	if NodeName == "" {
		f, err := os.ReadFile(overlayFile)
		if err != nil {
			wwlog.Error("Could not read file: %s", err)
			os.Exit(1)
		}

		fmt.Print(string(f))
	} else {
		if !util.IsFile(overlayFile) {
			wwlog.Debug("%s is not a file", overlayFile)
			wwlog.Error("%s:%s is not a file", overlayName, fileName)
			os.Exit(1)
		}
		if filepath.Ext(overlayFile) != ".ww" {
			wwlog.Warn("%s lacks the '.ww' suffix, will not be rendered in an overlay", fileName)
		}

		nodeDB, err := node.New()
		if err != nil {
			wwlog.Error("Could not open node configuration: %s", err)
			os.Exit(1)
		}
		nodes, err := nodeDB.FindAllNodes()
		if err != nil {
			wwlog.Error("Could not get node list: %s", err)
			os.Exit(1)
		}
		filteredNodes := node.FilterByName(nodes, []string{NodeName})
		if len(filteredNodes) != 1 {
			wwlog.Error("%v does not identify a single node", NodeName)
			os.Exit(1)
		}
		tstruct := overlay.InitStruct(&filteredNodes[0])
		tstruct.BuildSource = overlayFile
		buffer, backupFile, writeFile, err := overlay.RenderTemplateFile(overlayFile, tstruct)
		if err != nil {
			return err
		}
		var outBuffer bytes.Buffer
		// search for magic file name comment
		bufferScanner := bufio.NewScanner(bytes.NewReader(buffer.Bytes()))
		bufferScanner.Split(overlay.ScanLines)
		reg := regexp.MustCompile(`.*{{\s*/\*\s*file\s*["'](.*)["']\s*\*/\s*}}.*`)
		foundFileComment := false
		destFileName := strings.TrimSuffix(fileName, ".ww")
		for bufferScanner.Scan() {
			line := bufferScanner.Text()
			filenameFromTemplate := reg.FindAllStringSubmatch(line, -1)
			if len(filenameFromTemplate) != 0 {
				wwlog.Debug("Found multifile comment, new filename %s", filenameFromTemplate[0][1])
				if foundFileComment {
					if !Quiet {
						wwlog.Info("backupFile: %v\nwriteFile: %v", backupFile, writeFile)
						wwlog.Info("Filename: %s\n", destFileName)
					}
					wwlog.Info("%s", outBuffer.String())
					outBuffer.Reset()
				}
				destFileName = filenameFromTemplate[0][1]
				foundFileComment = true
			} else {
				_, _ = outBuffer.WriteString(line)
			}
		}
		if !Quiet {
			wwlog.Info("backupFile: %v\nwriteFile: %v", backupFile, writeFile)
			wwlog.Info("Filename: %s\n", destFileName)
		}
		fmt.Print(outBuffer.String())
	}
	return nil
}
