package reimprt

import (
	"github.com/containers/image/v5/signature"
	"github.com/spf13/cobra"
	apicontainer "github.com/warewulf/warewulf/internal/pkg/api/container"
	"github.com/warewulf/warewulf/internal/pkg/container"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		sCtx, err := apicontainer.GetSystemContext(vars.ociNoHttps, vars.ociUsername, vars.ociPassword)
		if err != nil {
			wwlog.ErrorExc(err, "")
			return
		}
		var pCtx *signature.PolicyContext
		pCtx, err = apicontainer.GetPolicyContext()
		if err != nil {
			_ = container.DeleteSource(args[1])
			return err
		}

		err = container.ReimportContainer(args[0], args[1], vars.recordChanges, sCtx, pCtx)
		if err != nil {
			return err
		}
		return nil
	}
}
