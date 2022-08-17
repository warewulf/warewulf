package set

import (
	"fmt"
	"os"

	apinode "github.com/hpcng/warewulf/internal/pkg/api/node"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/api/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	// OptionStrMap, netWithoutName := apinode.AddNetname(OptionStrMap)
	// if netWithoutName {
	// 	return errors.New("a netname must be given for any network related configuration")
	// }
	// realMap := make(map[string]string)

	// for key, val := range OptionStrMap {
	// 	realMap[key] = *val
	// }
	buffer, err := yaml.Marshal(nodeConf)
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
		var nodeCount uint
		// The checks run twice in the prompt case.
		// Avoiding putting in a blocking prompt in an API.
		_, nodeCount, err = apinode.NodeSetParameterCheck(&set, false)
		if err != nil {
			return
		}
		yes := util.ConfirmationPrompt(fmt.Sprintf("Are you sure you want to modify %d nodes(s)", nodeCount))
		if !yes {
			return
		}
	}
	return apinode.NodeSet(&set)
}
