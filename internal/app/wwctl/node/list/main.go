package list

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"strings"

	"github.com/hpcng/warewulf/internal/app/wwctl/helper"
	apinode "github.com/hpcng/warewulf/internal/pkg/api/node"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/util"
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

		var headers []string
		if vars.showAll {
			req.Type = wwapiv1.GetNodeList_All
			headers = []string{"NODE", "FIELD", "PROFILE", "VALUE"}
		} else if vars.showIpmi {
			req.Type = wwapiv1.GetNodeList_Ipmi
			headers = []string{"NODE NAME", "IPMI IPADDR", "IPMI PORT", "IPMI USERNAME", "IPMI INTERFACE", "IPMI ESCAPE CHAR"}
		} else if vars.showNet {
			req.Type = wwapiv1.GetNodeList_Network
			headers = []string{"NODE NAME", "NAME", "HWADDR", "IPADDR", "GATEWAY", "DEVICE"}
		} else if vars.showLong {
			req.Type = wwapiv1.GetNodeList_Long
			headers = []string{"NODE NAME", "KERNEL OVERRIDE", "CONTAINER", "OVERLAYS (S/R)"}
		} else if vars.showFullAll {
			req.Type = wwapiv1.GetNodeList_FullAll
			headers = []string{"NODE", "FIELD", "PROFILE", "VALUE"}
		} else {
			req.Type = wwapiv1.GetNodeList_Simple
			headers = []string{"NODE NAME", "PROFILES", "NETWORK"}
		}
		nodes, err := apinode.NodeList(&req)

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
			if err := csvWriter.Write(headers); err != nil {
				return err
			}
			for _, val := range nodes.Nodes {
				values := util.GetProtoMessageValues(val)
				if err := csvWriter.Write(values); err != nil {
					return err
				}
			}
		} else {
			ph := helper.NewPrintHelper(headers)
			for _, val := range nodes.Nodes {
				values := util.GetProtoMessageValues(val)
				ph.Append(values)
			}
			ph.Render()
		}
		return
	}
}
