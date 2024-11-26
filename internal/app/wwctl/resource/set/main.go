package set

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
			if nodeYml.Resource == nil {
				nodeYml.Resource = make(map[string]node.RemoteRes)
			}
			res := nodeYml.Resource[args[0]]
			for key, val := range vars.tags {
				res[key] = val
			}
			err = nodeYml.Persist()
			return err
		} else {
			return fmt.Errorf("resource %s does not exist", args[0])
		}
	}
}
