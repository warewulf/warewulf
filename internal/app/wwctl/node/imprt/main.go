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
	apiutil "github.com/warewulf/warewulf/internal/pkg/api/util"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
	"gopkg.in/yaml.v2"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	file, err := os.Open(args[0])
	if err != nil {
		wwlog.Error("Could not open file:%s \n", err)
		os.Exit(1)
	}
	defer file.Close()

	importMap := make(map[string]*node.NodeConf)
	buffer, err := io.ReadAll(file)
	if err != nil {
		wwlog.Error("Could not read:%s\n", err)
		os.Exit(1)
	}
	if !ImportCVS {
		err = yaml.Unmarshal(buffer, importMap)
		if err == nil {
			yes := apiutil.ConfirmationPrompt(fmt.Sprintf("Are you sure you want to modify %d nodes", len(importMap)))
			if yes {
				err = apinode.NodeAddFromYaml(&wwapiv1.NodeYaml{NodeConfMapYaml: string(buffer)})
				if err != nil {
					wwlog.Error("Got following problem when writing back yaml: %s", err)
					os.Exit(1)
				}
			}
		} else {
			wwlog.Error("Could not parse import file")
		}
	} else {
		// reading from buffer is a bit overshot
		csvReader := csv.NewReader(bytes.NewReader(buffer))
		records, err := csvReader.ReadAll()
		if err != nil {
			wwlog.Error("Could not parse %s: %s\n", args[0], err)
			os.Exit(1)
		}
		if len(records) < 1 || len(records[0]) < 1 {
			wwlog.Error("Did not find any data in %s\n", args[0])
			os.Exit(1)
		}
		if !(records[0][0] == "node" || records[0][0] == "nodename") {
			Usage()
			os.Exit(1)
		}
		argsLen := len(records[0])
		for i, line := range records[1:] {
			if len(line) != argsLen {
				wwlog.Error("Wrong number of fields in lube %u\n", i+1)
				os.Exit(1)
			}
			for j := range line {
				if j == 0 {
					continue
				}
				if importMap[line[0]] == nil {
					importMap[line[0]] = new(node.NodeConf)
				}
				ok := importMap[line[0]].SetLopt(records[0][j], line[j])
				if !(ok) {
					wwlog.Debug("Could not import %s\n", line[j])
				}
			}
		}
		yes := apiutil.ConfirmationPrompt(fmt.Sprintf("Are you sure you want to import %d nodes", len(importMap)))
		if yes {
			// create second buffer an marshall nodeMap to it
			buffer, err = yaml.Marshal(importMap)
			if err != nil {
				wwlog.Error("Got following problem when creating yaml: %s", err)
			}
			err = apinode.NodeAddFromYaml(&wwapiv1.NodeYaml{NodeConfMapYaml: string(buffer)})
			if err != nil {
				wwlog.Error("Got following problem when writing back yaml: %s", err)
				os.Exit(1)
			}
		}
	}

	return nil
}
