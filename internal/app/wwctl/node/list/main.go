package list

import (
	"strings"

	"github.com/hpcng/warewulf/internal/app/wwctl/helper"
	apinode "github.com/hpcng/warewulf/internal/pkg/api/node"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
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
		} else if vars.showYaml {
			req.Type = wwapiv1.GetNodeList_Yaml
		}
		nodeInfo, err := apinode.NodeList(&req)
		if vars.showYaml {
			wwlog.Output(strings.Join(nodeInfo.Output, " "))
			return
		}
		if len(nodeInfo.Output) > 0 {
			ph := helper.NewPrintHelper(strings.Split(nodeInfo.Output[0], ":=:"))
			for _, val := range nodeInfo.Output[1:] {
				ph.Append(strings.Split(val, ":=:"))
			}
			ph.Render()
		}
		return
	}
}
