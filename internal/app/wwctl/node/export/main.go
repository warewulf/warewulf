package export

import (
	"github.com/spf13/cobra"
	apinode "github.com/warewulf/warewulf/internal/pkg/api/node"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	filterList := wwapiv1.NodeList{
		Output: args,
	}
	nodeListMsg := apinode.FilteredNodes(&filterList)
	wwlog.Info(nodeListMsg.NodeConfMapYaml)
	return nil
}
