package version

import (
	"fmt"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
)

/*
Return the version of wwctl
*/
func GetVersion() string {
	return fmt.Sprintf("%s-%s", buildconfig.VERSION(), buildconfig.RELEASE())
}

/*
Returns the version of the api via grpc
*/
func Version() (versionResponse *wwapiv1.VersionResponse) {
	versionResponse = &wwapiv1.VersionResponse{}
	versionResponse.ApiPrefix = "rc1"
	versionResponse.ApiVersion = "1"
	versionResponse.WarewulfVersion = GetVersion()
	return
}
