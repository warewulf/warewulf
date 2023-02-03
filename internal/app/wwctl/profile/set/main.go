package set

import (
	"fmt"
	"os"

	apiprofile "github.com/hpcng/warewulf/internal/pkg/api/profile"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/api/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {

	// remove the default network as the all network values are assigned
	// to this network
	if NetName != "default" {
		ProfileConf.NetDevs[NetName] = ProfileConf.NetDevs["default"]
		delete(ProfileConf.NetDevs, "default")
	}
	buffer, err := yaml.Marshal(ProfileConf)
	if err != nil {
		wwlog.Error("Cant marshall nodeInfo", err)
		os.Exit(1)
	}
	set := wwapiv1.NodeSetParameter{
		NodeConfYaml: string(buffer[:]),
		NetdevDelete: SetNetDevDel,
		AllNodes:     SetNodeAll,
		Force:        SetForce,
		NodeNames:    args,
	}

	if !SetYes {
		var profileCount uint
		// The checks run twice in the prompt case.
		// Avoiding putting in a blocking prompt in an API.
		_, profileCount, err = apiprofile.ProfileSetParameterCheck(&set, false)
		if err != nil {
			return
		}
		yes := util.ConfirmationPrompt(fmt.Sprintf("Are you sure you want to modify %d profile(s)", profileCount))
		if !yes {
			return
		}
	}
	return apiprofile.ProfileSet(&set)
}
