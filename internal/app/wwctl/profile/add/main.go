package add

import (
	"fmt"

	apinode "github.com/hpcng/warewulf/internal/pkg/api/node"
	apiprofile "github.com/hpcng/warewulf/internal/pkg/api/profile"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/api/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	OptionStrMap, haveNetname := apinode.AddNetname(OptionStrMap)
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
		// The checks run twice in the prompt case.
		// Avoiding putting in a blocking prompt in an API.
		apiprofile.AddProfile(&set, false)
		if err != nil {
			return
		}
		_, _, err = apiprofile.ProfileSetParameterCheck(&set, false)
		if err != nil {
			return
		}

		yes := util.ConfirmationPrompt(fmt.Sprintf("Are you sure you add the profile %s", args))
		if !yes {
			return
		}
	}
	return apiprofile.ProfileSet(&set)
}
