package add

import (
	"os"

	"gopkg.in/yaml.v2"

	apinode "github.com/hpcng/warewulf/internal/pkg/api/node"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
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
	set := wwapiv1.NodeAddParameter{
		NodeConfYaml: string(buffer[:]),
		NodeNames:    args,
	}

	return apinode.NodeAdd(&set)
}
