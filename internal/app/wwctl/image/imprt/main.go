package imprt

import (
	"github.com/spf13/cobra"
	apiimage "github.com/warewulf/warewulf/internal/pkg/api/image"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {

	// Shim in a name if none given.
	name := ""
	if len(args) == 2 {
		name = args[1]
	}

	cip := &wwapiv1.ImageImportParameter{
		Source:      args[0],
		Name:        name,
		Force:       SetForce,
		Update:      SetUpdate,
		Build:       SetBuild,
		SyncUser:    SyncUser,
		OciNoHttps:  OciNoHttps,
		OciUsername: OciUsername,
		OciPassword: OciPassword,
		Platform:    Platform,
	}

	_, err = apiimage.ImageImport(cip)
	return
}
