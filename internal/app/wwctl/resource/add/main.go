package add

import (
	"fmt"

	"github.com/warewulf/warewulf/internal/pkg/node"

	"github.com/spf13/cobra"
)

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		nodeYml, err := node.New()
		if err != nil {
			return fmt.Errorf("failed to load node configuration: %s", err)
		}
		if ok := nodeYml.Resource[args[0]] != nil; ok {
			return fmt.Errorf("resource %s already exists", args[0])
		}
		if nodeYml.Resource == nil {
			nodeYml.Resource = make(map[string]node.RemoteRes)
		}
		res := node.RemoteRes{}
		for key, val := range vars.tags {
			res[key] = val
		}
		nodeYml.Resource[args[0]] = res
		err = nodeYml.Persist()
		return err
	}
}
