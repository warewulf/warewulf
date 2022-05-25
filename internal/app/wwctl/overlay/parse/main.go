package parse

import (
	"bufio"
	"bytes"
	"os"
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
	var host node.NodeInfo
	if NodeName == "" {
		host.Kernel = new(node.KernelEntry)
		host.Ipmi = new(node.IpmiEntry)
		var idEntry node.Entry
		hostname, _ := os.Hostname()
		idEntry.Set(hostname)
		host.Id = idEntry
	} else {
		nodeDB, err := node.New()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
			os.Exit(1)
		}
		nodes, err := nodeDB.FindAllNodes()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not get node list: %s\n", err)
			os.Exit(1)
		}
		node := node.FilterByName(nodes, []string{NodeName})
		if len(node) != 1 {
			wwlog.Printf(wwlog.ERROR, "%v does not identify a single node\n", NodeName)
			os.Exit(1)
		}
		host = node[0]
	}
	if !util.IsFile(args[0]) {
		wwlog.Printf(wwlog.ERROR, "%s is not a file\n", args[0])
	}
	tstruct := overlay.InitStruct(host)
	tstruct.BuildSource = args[0]
	buffer, backupFile, writeFile, err := overlay.RenderTemplateFile(args[0], tstruct)
	if err != nil {
		return err
	}
	if filepath.Ext(args[0]) != ".ww" {
		wwlog.Printf(wwlog.WARN, "%s has not the '.ww' so wont be rendered if in overlay\n", args[0])
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
			wwlog.Printf(wwlog.DEBUG, "Found multifile comment, new filename %s\n", filenameFromTemplate[0][1])
			if foundFileComment {
				wwlog.Printf(wwlog.INFO, "backupFile: %v\nwriteFile: %v\n", backupFile, writeFile)
				wwlog.Printf(wwlog.INFO, "Filename: %s\n\n", destFileName)
				wwlog.Printf(wwlog.INFO, "%s", outBuffer.String())
				outBuffer.Reset()
			}
			destFileName = filenameFromTemplate[0][1]
			foundFileComment = true
		} else {
			_, _ = outBuffer.WriteString(line)
		}
	}
	wwlog.Printf(wwlog.INFO, "backupFile: %v\nwriteFile: %v\n", backupFile, writeFile)
	wwlog.Printf(wwlog.INFO, "Filename: %s\n\n", destFileName)
	wwlog.Printf(wwlog.INFO, "%s", outBuffer.String())
	return nil
}
