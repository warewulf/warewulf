package build

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/container"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	var containers []string
	if BuildAll {
		containers, err = container.ListSources()
		if err != nil {
			return fmt.Errorf("failed to list all containers")
		}
	} else {
		containers = args
	}
	return container.Build(&container.BuildParameter{
		Names:   containers,
		Force:   BuildForce,
		Default: SetDefault,
	})
}
