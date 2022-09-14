package show

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
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
		wwlog.Error("Overlay does not exist: %s\n", overlayName)
		os.Exit(1)
	}

	overlayFile := path.Join(overlaySourceDir, fileName)

	if !util.IsFile(overlayFile) {
		wwlog.Error("File does not exist within overlay: %s:%s\n", overlayName, fileName)
		os.Exit(1)
	}

	if NodeName == "" {
		f, err := ioutil.ReadFile(overlayFile)
		if err != nil {
			wwlog.Error("Could not read file: %s\n", err)
			os.Exit(1)
		}

		fmt.Print(string(f))
	} else {
		var host node.NodeInfo
		nodeDB, err := node.New()
		if err != nil {
			wwlog.Error("Could not open node configuration: %s\n", err)
			os.Exit(1)
		}
		nodes, err := nodeDB.FindAllNodes()
		if err != nil {
			wwlog.Error("Could not get node list: %s\n", err)
			os.Exit(1)
		}
		node := node.FilterByName(nodes, []string{NodeName})
		if len(node) != 1 {
			wwlog.Error("%v does not identify a single node\n", NodeName)
			os.Exit(1)
		}
		host = node[0]

		if !util.IsFile(args[0]) {
			wwlog.Error("%s is not a file\n", args[0])
		}
		tstruct := overlay.InitStruct(host)
		tstruct.BuildSource = args[0]
		buffer, backupFile, writeFile, err := overlay.RenderTemplateFile(overlayFile, tstruct)
		if err != nil {
			return err
		}
		if filepath.Ext(args[0]) != ".ww" {
			wwlog.Warn("%s has not the '.ww' so wont be rendered if in overlay\n", args[0])
		}
		var outBuffer bytes.Buffer
		// search for magic file name comment
		bufferScanner := bufio.NewScanner(bytes.NewReader(buffer.Bytes()))
		bufferScanner.Split(overlay.ScanLines)
		reg := regexp.MustCompile(`.*{{\s*/\*\s*file\s*["'](.*)["']\s*\*/\s*}}.*`)
		foundFileComment := false
		destFileName := strings.TrimSuffix(args[0], ".ww")
		for bufferScanner.Scan() {
			line := bufferScanner.Text()
			filenameFromTemplate := reg.FindAllStringSubmatch(line, -1)
			if len(filenameFromTemplate) != 0 {
				wwlog.Debug("Found multifile comment, new filename %s\n", filenameFromTemplate[0][1])
				if foundFileComment {
					if !Quiet {
						wwlog.Info("backupFile: %v\nwriteFile: %v\n", backupFile, writeFile)
						wwlog.Info("Filename: %s\n\n", destFileName)
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
			wwlog.Info("backupFile: %v\nwriteFile: %v\n", backupFile, writeFile)
			wwlog.Info("Filename: %s\n\n", destFileName)
		}
		wwlog.Info("%s", outBuffer.String())

	}
	return nil
}
