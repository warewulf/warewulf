package add

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/node"
)

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		// remove the UNDEF network as all network values are assigned
		// to this network
		if !node.ObjectIsEmpty(vars.profileConf.NetDevs["UNDEF"]) || len(vars.profileAdd.NetTagsAdd) > 0 {
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
		nodeDB, err := node.New()
		if err != nil {
			return fmt.Errorf("could not open database: %w", err)
		}
		changed := cmd.Flags().Changed
		for _, p := range args {
			if slices.Contains(nodeDB.ListAllProfiles(), p) {
				return fmt.Errorf("profile with name %s already exists", p)
			}
			pNew, err := nodeDB.AddProfile(p)
			if err != nil {
				return err
			}
			pNew.UpdateFrom(&vars.profileConf, changed)
			pNew.ApplyTagChanges(vars.profileAdd, node.NodeConfDel{})
		}
		return nodeDB.Persist()
	}
}
