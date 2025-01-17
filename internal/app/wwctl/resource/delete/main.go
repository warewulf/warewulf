package delete

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
		for _, res := range args {
			if _, ok := nodeYml.Resources[res]; !ok {
				return fmt.Errorf("resource %s does not exist", res)
			}
			delete(nodeYml.Resources, res)
		}
		return nodeYml.Persist()
	}
}
