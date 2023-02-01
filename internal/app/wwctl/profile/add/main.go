package add

import (
	"fmt"
	"os"

	apiprofile "github.com/hpcng/warewulf/internal/pkg/api/profile"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"gopkg.in/yaml.v2"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/api/util"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	// remove the default network as the all network values are assigned
	// to this network
	if NetName != "" {
		netDev := *ProfileConf.NetDevs["default"]
		ProfileConf.NetDevs[NetName] = &netDev
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
		// The checks run twice in the prompt case.
		// Avoiding putting in a blocking prompt in an API.
		_, _, err = apiprofile.ProfileSetParameterCheck(&set, false)
		if err != nil {
			return
		}

		yes := util.ConfirmationPrompt(fmt.Sprintf("Are you sure you add the profile %s", args))
		if !yes {
			return
		}
	}
	return apiprofile.AddProfile(&set, false)
}
