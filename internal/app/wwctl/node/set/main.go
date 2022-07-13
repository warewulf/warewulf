package set

import (
	"errors"
	"fmt"

	"github.com/hpcng/warewulf/internal/pkg/api/node"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/api/util"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	OptionStrMap, haveNetname := node.AddNetname(OptionStrMap)
	if !haveNetname {
		return errors.New("a netname must be given for any network related configuration")
	}
	realMap := make(map[string]string)

	for key, val := range OptionStrMap {
		realMap[key] = *val
	}

	set := wwapiv1.NodeSetParameter{
		OptionsStrMap: realMap,
		NetdevDelete:  SetNetDevDel,
		AllNodes:      SetNodeAll,
		Force:         SetForce,
		NodeNames:     args,
	}

	if !SetYes {
		var nodeCount uint
		// The checks run twice in the prompt case.
		// Avoiding putting in a blocking prompt in an API.
		_, nodeCount, err = node.NodeSetParameterCheck(&set, false)
		if err != nil {
			return
		}
		yes := util.ConfirmationPrompt(fmt.Sprintf("Are you sure you want to modify %d nodes(s)", nodeCount))
		if !yes {
			return
		}
	}
	return node.NodeSet(&set)
}
