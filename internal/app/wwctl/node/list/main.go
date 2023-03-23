package list

import (
	"strings"

	"github.com/hpcng/warewulf/internal/app/wwctl/helper"
	apinode "github.com/hpcng/warewulf/internal/pkg/api/node"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	req := wwapiv1.GetNodeList{
		Nodes: args,
		Type:  wwapiv1.GetNodeList_Simple,
	}
	if ShowAll {
		req.Type = wwapiv1.GetNodeList_All
	} else if ShowIpmi {
		req.Type = wwapiv1.GetNodeList_Ipmi
	} else if ShowNet {
		req.Type = wwapiv1.GetNodeList_Network
	} else if ShowLong {
		req.Type = wwapiv1.GetNodeList_Long
	}
	nodeInfo, err := apinode.NodeList(&req)
	if len(nodeInfo.Output) > 0 {
		ph := helper.NewPrintHelper(strings.Split(nodeInfo.Output[0], "="))
		for _, val := range nodeInfo.Output[1:] {
			ph.Append(strings.Split(val, "="))
		}
		ph.Render()
	}
	return
}
