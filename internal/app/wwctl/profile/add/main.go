package add

import (
	"fmt"
	"strings"

	apiprofile "github.com/warewulf/warewulf/internal/pkg/api/profile"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
)

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		// remove the UNDEF network as all network values are assigned
		// to this network
		if !node.ObjectIsEmpty(vars.profileConf.NetDevs["UNDEF"]) {
			netDev := *vars.profileConf.NetDevs["UNDEF"]
			vars.profileConf.NetDevs[vars.nodeAdd.Net] = &netDev
		}
		delete(vars.profileConf.NetDevs, "UNDEF")
		if vars.nodeAdd.FsName != "" {
			if !strings.HasPrefix(vars.nodeAdd.FsName, "/dev") {
				if vars.nodeAdd.FsName == vars.nodeAdd.PartName {
					vars.nodeAdd.FsName = "/dev/disk/by-partlabel/" + vars.nodeAdd.PartName
				} else {
					return fmt.Errorf("filesystems need to have a underlying blockdev")
				}
			}
			fs := *vars.profileConf.FileSystems["UNDEF"]
			vars.profileConf.FileSystems[vars.nodeAdd.FsName] = &fs
		}
		delete(vars.profileConf.FileSystems, "UNDEF")
		if vars.nodeAdd.DiskName != "" && vars.nodeAdd.PartName != "" {
			prt := *vars.profileConf.Disks["UNDEF"].Partitions["UNDEF"]
			vars.profileConf.Disks["UNDEF"].Partitions[vars.nodeAdd.PartName] = &prt
			delete(vars.profileConf.Disks["UNDEF"].Partitions, "UNDEF")
			dsk := *vars.profileConf.Disks["UNDEF"]
			vars.profileConf.Disks[vars.nodeAdd.DiskName] = &dsk
		}
		if (vars.nodeAdd.DiskName != "") != (vars.nodeAdd.PartName != "") {
			return fmt.Errorf("partition and disk must be specified")
		}
		delete(vars.profileConf.Disks, "UNDEF")
		buffer, err := yaml.Marshal(vars.profileConf)
		if err != nil {
			return fmt.Errorf("can not marshall nodeInfo: %w", err)
		}
		set := wwapiv1.NodeAddParameter{
			NodeConfYaml: string(buffer[:]),
			NodeNames:    args,
			Force:        true,
		}
		return apiprofile.ProfileAdd(&set)
	}
}
