package add

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
)

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		// remove the UNDEF network as all network values are assigned
		// to this network
		if !node.ObjectIsEmpty(vars.profileConf.NetDevs["UNDEF"]) {
			netDev := *vars.profileConf.NetDevs["UNDEF"]
			vars.profileConf.NetDevs[vars.profileAdd.Net] = &netDev
		}
		delete(vars.profileConf.NetDevs, "UNDEF")
		if vars.profileAdd.FsName != "" {
			if !strings.HasPrefix(vars.profileAdd.FsName, "/dev") {
				if vars.profileAdd.FsName == vars.profileAdd.PartName {
					vars.profileAdd.FsName = "/dev/disk/by-partlabel/" + vars.profileAdd.PartName
				} else {
					return fmt.Errorf("filesystems need to have a underlying blockdev")
				}
			}
			fs := *vars.profileConf.FileSystems["UNDEF"]
			vars.profileConf.FileSystems[vars.profileAdd.FsName] = &fs
		}
		delete(vars.profileConf.FileSystems, "UNDEF")
		if vars.profileAdd.DiskName != "" && vars.profileAdd.PartName != "" {
			prt := *vars.profileConf.Disks["UNDEF"].Partitions["UNDEF"]
			vars.profileConf.Disks["UNDEF"].Partitions[vars.profileAdd.PartName] = &prt
			delete(vars.profileConf.Disks["UNDEF"].Partitions, "UNDEF")
			dsk := *vars.profileConf.Disks["UNDEF"]
			vars.profileConf.Disks[vars.profileAdd.DiskName] = &dsk
		}
		if (vars.profileAdd.DiskName != "") != (vars.profileAdd.PartName != "") {
			return fmt.Errorf("partition and disk must be specified")
		}
		delete(vars.profileConf.Disks, "UNDEF")
		vars.profileConf.Ipmi.Tags = vars.profileAdd.IpmiTagsAdd
		buffer, err := yaml.Marshal(vars.profileConf)
		if err != nil {
			return fmt.Errorf("can not marshall nodeInfo: %w", err)
		}
		nodeDB, err := node.New()
		if err != nil {
			return fmt.Errorf("could not open database: %w", err)
		}
		for _, p := range args {
			if util.InSlice(nodeDB.ListAllProfiles(), p) {
				return fmt.Errorf("profile with name %s already exists", p)
			}
			pNew, err := nodeDB.AddProfile(p)
			if err != nil {
				return err
			}
			err = yaml.Unmarshal(buffer, pNew)
			if err != nil {
				return fmt.Errorf("failed to add profile: %w", err)
			}
		}
		return nodeDB.Persist()
	}
}
