package export

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	registry, err := node.New()
	if err != nil {
		return err
	}
	nodeMap := make(map[string]*node.Node)
	names := hostlist.Expand(args)
	if len(names) == 0 {
		names = registry.ListAllNodes()
	}
	for _, name := range hostlist.Expand(names) {
		if n, err := registry.GetNode(name); err == nil {
			nodeMap[name] = &n
		}
	}
	y, err := util.EncodeYaml(nodeMap)
	if err != nil {
		return err
	}
	wwlog.Output("%s", y)
	return nil
}
