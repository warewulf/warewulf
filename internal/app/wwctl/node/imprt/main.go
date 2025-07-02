package imprt

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	apinode "github.com/warewulf/warewulf/internal/pkg/api/node"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
	"gopkg.in/yaml.v3"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	file, err := os.Open(args[0])
	if err != nil {
		return fmt.Errorf("could not open file: %s", err)
	}
	defer file.Close()

	importMap := make(map[string]*node.Node)
	buffer, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("could not read: %s", err)
	}
	if !ImportCSV {
		err = yaml.Unmarshal(buffer, importMap)
		if err == nil {
			if setYes || util.Confirm(fmt.Sprintf("Are you sure you want to modify %d nodes", len(importMap))) {
				err = apinode.NodeAddFromYaml(&wwapiv1.NodeYaml{NodeConfMapYaml: string(buffer)})
				if err != nil {
					return fmt.Errorf("got following problem when writing back yaml: %s", err)
				}
			}
		} else {
			return fmt.Errorf("could not parse import file: %s", err)
		}
	} else {
		// reading from buffer is a bit overshot
		csvReader := csv.NewReader(bytes.NewReader(buffer))
		records, err := csvReader.ReadAll()
		if err != nil {
			return fmt.Errorf("could not parse %s: %s", args[0], err)
		}
		if len(records) < 1 || len(records[0]) < 1 {
			return fmt.Errorf("did not find any data in %s", args[0])
		}
		if !(records[0][0] == "node" || records[0][0] == "nodename") {
			Usage()
			os.Exit(1)
		}
		fieldNames := records[0]
		argsLen := len(fieldNames)
		for i, line := range records[1:] {
			if len(line) != argsLen {
				return fmt.Errorf("wrong number of fields in lube %d", i+1)
			}
			for j := range line {
				if j == 0 {
					continue
				}
				nodeName := line[0]
				fieldName := fieldNames[j]
				fieldValue := line[j]
				if importMap[nodeName] == nil {
					importMap[nodeName] = new(node.Node)
				}
				ok := importMap[nodeName].SetLopt(fieldName, fieldValue)
				if !(ok) {
					wwlog.Debug("Could not import %s", fieldValue)
				}
			}
		}
		if setYes || util.Confirm(fmt.Sprintf("Are you sure you want to import %d nodes", len(importMap))) {
			// create second buffer an marshall nodeMap to it
			buffer, err = yaml.Marshal(importMap)
			if err != nil {
				return fmt.Errorf("got following problem when creating yaml: %s", err)
			}
			err = apinode.NodeAddFromYaml(&wwapiv1.NodeYaml{NodeConfMapYaml: string(buffer)})
			if err != nil {
				return fmt.Errorf("got following problem when writing back yaml: %s", err)
			}
		}
	}

	return nil
}
