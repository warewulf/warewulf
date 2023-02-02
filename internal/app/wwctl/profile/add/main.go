package add

import (
	"os"

	"gopkg.in/yaml.v2"

	apiprofile "github.com/hpcng/warewulf/internal/pkg/api/profile"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	// remove the default network as the all network values are assigned
	// to this network
	if NetName != "" && NetName != "default" {
		ProfileConf.NetDevs[NetName] = ProfileConf.NetDevs["default"]
		delete(ProfileConf.NetDevs, "default")
	}
	buffer, err := yaml.Marshal(ProfileConf)
	if err != nil {
		wwlog.Error("Can't marshall profile configuration", err)
		os.Exit(1)
	}
	wwlog.Debug("profile add buffer: %s", buffer)
	set := wwapiv1.NodeAddParameter{
		NodeConfYaml: string(buffer[:]),
		NodeNames:    args,
	}

	return apiprofile.ProfileAdd(&set)
}
