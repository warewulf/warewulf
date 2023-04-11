package add

import (
	"gopkg.in/yaml.v2"

	apinode "github.com/hpcng/warewulf/internal/pkg/api/node"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

/*
RunE needs a function of type func(*cobraCommand,[]string) err, but
in order to avoid global variables which mess up testing a function of
the required type is returned
*/
func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// run converters for different types
		for _, c := range vars.converters {
			if err := c(); err != nil {
				return err
			}
		}
		// remove the default network as all network values are assigned
		// to this network
		if _, ok := vars.nodeConf.NetDevs["default"]; ok && vars.netName != "" {
			netDev := *vars.nodeConf.NetDevs["default"]
			vars.nodeConf.NetDevs[vars.netName] = &netDev
			delete(vars.nodeConf.NetDevs, "default")
		} else {
			if vars.nodeConf.NetDevs["default"].Empty() {
				delete(vars.nodeConf.NetDevs, "default")
			}
		}
		buffer, err := yaml.Marshal(vars.nodeConf)
		if err != nil {
			wwlog.Error("Cant marshall nodeInfo", err)
			return err
		}
		set := wwapiv1.NodeAddParameter{
			NodeConfYaml: string(buffer[:]),
			NodeNames:    args,
			Force:        true,
		}
		return apinode.NodeAdd(&set)
	}
}
