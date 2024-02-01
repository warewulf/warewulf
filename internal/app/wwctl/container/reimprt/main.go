package reimprt

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/containers/image/v5/types"
	"github.com/spf13/cobra"
	apicontainer "github.com/warewulf/warewulf/internal/pkg/api/container"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/warewulf/warewulf/internal/pkg/container"
	"github.com/warewulf/warewulf/internal/pkg/oci"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		src := args[0]
		if !container.ValidSource(src) {
			return fmt.Errorf("container doesn't exist: %s", src)
		}
		inspectSrc := oci.InspectOutput{}
		inspectFile, err := os.Open(path.Join(container.SourceDir(src), "src/inspect.json"))
		if err != nil {
			return fmt.Errorf("couldn't open inspect data of source container, reimport isn't possible: %s", err)
		}
		defer inspectFile.Close()
		buf, _ := io.ReadAll(inspectFile)
		err = json.Unmarshal(buf, &inspectSrc)
		if err != nil {
			return fmt.Errorf("couldn't unmarshall inspect data for container source: %s", err)
		}
		if !vars.fromCache {
			cip := &wwapiv1.ContainerImportParameter{
				Source:      "docker://" + inspectSrc.Name,
				Name:        args[1],
				SyncUser:    vars.syncUser,
				Build:       vars.setBuild,
				OciNoHttps:  vars.ociNoHttps,
				OciUsername: vars.ociUsername,
				OciPassword: vars.ociPassword,
			}

			_, err = apicontainer.ContainerImport(cip)
			return err
		} else {
			var sCtx *types.SystemContext
			sCtx, err = apicontainer.GetSystemContext(vars.ociNoHttps, vars.ociUsername, vars.ociPassword)
			if err != nil {
				wwlog.ErrorExc(err, "")
				return
			}
			err = container.ReimportContainer(inspectSrc, args[1], sCtx)
			if err != nil {
				return err
			}

		}
		return nil
	}
}
