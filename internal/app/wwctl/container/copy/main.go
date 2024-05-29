package copy

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/container"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {

	if len(args) > 2 {
		wwlog.Warn("copy only requires 2 arguments but you provided %d arguments. Hence, they will be ignored.", len(args))
	}

	err = container.Copy(&container.CopyParameter{
		Name:        args[0],
		Destination: args[1],
		ForceBuild:  true,
	})
	if err != nil {
		err = fmt.Errorf("could not duplicate image: %s", err.Error())
		wwlog.Error(err.Error())
		return
	}

	wwlog.Info("Container %s successfully duplicated as %s", args[0], args[1])
	return

}
