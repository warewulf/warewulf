package version

import (
	"fmt"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
)

/*
Return the version of wwctl
*/
func GetVersion() string {
	conf := warewulfconf.New()
	return fmt.Sprintf("%s-%s", conf.Paths.Version, conf.Paths.Release)
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
