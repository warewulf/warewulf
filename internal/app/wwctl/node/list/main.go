package list

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"sort"
	"strings"

	"github.com/hpcng/warewulf/internal/app/wwctl/helper"
	apinode "github.com/hpcng/warewulf/internal/pkg/api/node"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		if len(args) > 0 && strings.Contains(args[0], ",") {
			args = strings.FieldsFunc(args[0], func(r rune) bool { return r == ',' })
		}
		req := wwapiv1.GetNodeList{
			Nodes: args,
			Type:  wwapiv1.GetNodeList_Simple,
		}

		if vars.showAll {
			req.Type = wwapiv1.GetNodeList_All
		} else if vars.showIpmi {
			req.Type = wwapiv1.GetNodeList_Ipmi
		} else if vars.showNet {
			req.Type = wwapiv1.GetNodeList_Network
		} else if vars.showLong {
			req.Type = wwapiv1.GetNodeList_Long
		} else if vars.showFullAll {
			req.Type = wwapiv1.GetNodeList_FullAll
		} else {
			req.Type = wwapiv1.GetNodeList_Simple
		}
		nodes, err := apinode.NodeList(&req)

		if len(nodes.Nodes) > 0 {
			if strings.EqualFold(strings.TrimSpace(vars.output), "yaml") {
				yamlBytes, err := yaml.Marshal(nodes)
				if err != nil {
					return err
				}

				wwlog.Info(string(yamlBytes))
			} else if strings.EqualFold(strings.TrimSpace(vars.output), "json") {
				jsonBytes, err := json.Marshal(nodes)
				if err != nil {
					return err
				}

				wwlog.Info(string(jsonBytes))
			} else if strings.EqualFold(strings.TrimSpace(vars.output), "csv") {
				csvWriter := csv.NewWriter(os.Stdout)
				defer csvWriter.Flush()

				headerWrite := false

				keys := make([]string, 0, len(nodes.Nodes))
				for k := range nodes.Nodes {
					keys = append(keys, k)
				}
				sort.Strings(keys)

				for _, key := range keys {
					vals := nodes.Nodes[key]
					if !headerWrite {
						if err := csvWriter.Write(vals[0].GetHeader()); err != nil {
							return err
						}
						headerWrite = true
					}

					for _, val := range vals {
						columns := []string{key}
						columns = append(columns, val.GetValue()...)
						if err := csvWriter.Write(columns); err != nil {
							return err
						}
					}
				}
			} else {
				var ph *helper.PrintHelper
				headerWrite := false

				keys := make([]string, 0, len(nodes.Nodes))
				for k := range nodes.Nodes {
					keys = append(keys, k)
				}
				sort.Strings(keys)

				for _, key := range keys {
					vals := nodes.Nodes[key]
					if !headerWrite {
						ph = helper.NewPrintHelper(vals[0].GetHeader())
					}
					headerWrite = true
					for _, val := range vals {
						columns := []string{key}
						columns = append(columns, val.GetValue()...)
						ph.Append(columns)
					}
				}
				ph.Render()
			}
		}

		return
	}
}
