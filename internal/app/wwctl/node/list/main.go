package list

import (
	"fmt"

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
	for _, str := range nodeInfo.Output {
		fmt.Printf("%s\n", str)
	}
	return
}
