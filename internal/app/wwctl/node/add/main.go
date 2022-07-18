package add

import (
	"errors"

	apinode "github.com/hpcng/warewulf/internal/pkg/api/node"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	OptionStrMap, haveNetname := apinode.AddNetname(OptionStrMap)
	if !haveNetname {
		return errors.New("a netname must be given for any network related configuration")
	}
	realMap := make(map[string]string)

	for key, val := range OptionStrMap {
		realMap[key] = *val
	}
	set := wwapiv1.NodeAddParameter{
		OptionsStrMap: realMap,
		NodeNames:     args,
	}

	return apinode.NodeAdd(&set)
}
