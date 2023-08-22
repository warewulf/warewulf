package add

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"

	apinode "github.com/hpcng/warewulf/internal/pkg/api/node"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/node"
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
		// remove the UNDEF network as all network values are assigned
		// to this network
		if !node.ObjectIsEmpty(vars.nodeConf.NetDevs["UNDEF"]) {
			netDev := *vars.nodeConf.NetDevs["UNDEF"]
			vars.nodeConf.NetDevs[vars.netName] = &netDev
		}
		delete(vars.nodeConf.NetDevs, "UNDEF")
		if vars.fsName != "" {
			if !strings.HasPrefix(vars.fsName, "/dev") {
				if vars.fsName == vars.partName {
					vars.fsName = "/dev/disk/by-partlabel/" + vars.partName
				} else {
					return fmt.Errorf("filesystems need to have a underlying blockdev")
				}
			}
			fs := *vars.nodeConf.FileSystems["UNDEF"]
			vars.nodeConf.FileSystems[vars.fsName] = &fs
		}
		delete(vars.nodeConf.FileSystems, "UNDEF")
		if vars.diskName != "" && vars.partName != "" {
			prt := *vars.nodeConf.Disks["UNDEF"].Partitions["UNDEF"]
			vars.nodeConf.Disks["UNDEF"].Partitions[vars.partName] = &prt
			delete(vars.nodeConf.Disks["UNDEF"].Partitions, "UNDEF")
			dsk := *vars.nodeConf.Disks["UNDEF"]
			vars.nodeConf.Disks[vars.diskName] = &dsk
		}
		if (vars.diskName != "") != (vars.partName != "") {
			return fmt.Errorf("partition and disk must be specified")
		}
		delete(vars.nodeConf.Disks, "UNDEF")
		buffer, err := yaml.Marshal(vars.nodeConf)
		if err != nil {
			wwlog.Error("Can't marshall nodeInfo", err)
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
