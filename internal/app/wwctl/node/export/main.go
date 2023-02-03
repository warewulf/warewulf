package export

import (
	"fmt"

	apinode "github.com/hpcng/warewulf/internal/pkg/api/node"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		args = append(args, ".*")
	}
	filterList := wwapiv1.NodeList{
		Output: args,
	}
	nodeListMsg := apinode.FilteredNodes(&filterList)
	/*
		nodeMap := make(map[string]*node.NodeConf)
		// got proper yaml back
		_ = yaml.Unmarshal([]byte(nodeListMsg.NodeConfMapYaml), nodeMap)
	*/
	fmt.Println(nodeListMsg.NodeConfMapYaml)
	return nil
}
