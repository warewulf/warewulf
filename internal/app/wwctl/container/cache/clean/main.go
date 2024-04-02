package cacheclean

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/container"
)

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {

		for _, arg := range args {
			err = container.DeleteCache(arg)
			if err != nil {
				return err
			}
		}
		return
	}
}
