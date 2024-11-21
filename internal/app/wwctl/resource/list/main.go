package list

import (
	"fmt"

	"github.com/warewulf/warewulf/internal/app/wwctl/table"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"

	"github.com/spf13/cobra"
)

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		nodeYml, err := node.New()
		if err != nil {
			return fmt.Errorf("failed to load node configuration: %s", err)
		}
		var resList []string
		if len(args) == 1 {
			resList = []string{args[0]}
		} else {
			resList = nodeYml.ListAllResources()
		}
		for _, resname := range resList {
			res, err := nodeYml.GetResource(resname)
			if err != nil {
				return err
			}
			if !vars.all {
				wwlog.Info("%s", resname)
			} else {
				t := table.New(cmd.OutOrStdout())
				t.AddHeader(resname, "key", "value")
				for key, val := range res {
					t.AddLine(resname, key, val)
				}
				t.Print()

			}
		}
		return nil
	}
}
