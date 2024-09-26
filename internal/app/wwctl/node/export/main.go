package export

import (
	"github.com/spf13/cobra"
	apinode "github.com/warewulf/warewulf/internal/pkg/api/node"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
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
	wwlog.Info(nodeListMsg.NodeConfMapYaml)
	return nil
}
