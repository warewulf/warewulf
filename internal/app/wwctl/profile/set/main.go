package set

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) error {

		// remove the default network as the all network values are assigned
		// to this network
		if !node.ObjectIsEmpty(vars.profileConf.NetDevs["UNDEF"]) || len(vars.profileAdd.NetTagsAdd) > 0 {
			netDev := *vars.profileConf.NetDevs["UNDEF"]
			vars.profileConf.NetDevs[vars.profileAdd.Net] = &netDev
			vars.profileConf.NetDevs[vars.profileAdd.Net].Tags = vars.profileAdd.NetTagsAdd
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

		nodeDB, err := node.New()
		if err != nil {
			return fmt.Errorf("could not open configuration: %w", err)
		}

		if len(args) == 0 {
			return fmt.Errorf("no profiles specified")
		} else if len(nodeDB.ListAllProfiles()) == 0 {
			wwlog.Warn("no nodes/profiles found")
			return nil
		}

		changed := cmd.Flags().Changed
		var count uint
		for _, profileId := range args {
			wwlog.Verbose("evaluating profile: %s", profileId)
			profilePtr, err := nodeDB.GetProfilePtr(profileId)
			if err != nil {
				wwlog.Warn("invalid profile: %s", profileId)
				continue
			}
			profilePtr.UpdateFrom(&vars.profileConf, changed)
			if vars.profileDel.NetDel != "" {
				if _, ok := profilePtr.NetDevs[vars.profileDel.NetDel]; !ok {
					return fmt.Errorf("network device name doesn't exist: %s", vars.profileDel.NetDel)
				}
				wwlog.Verbose("Profile: %s, Deleting network device: %s", profileId, vars.profileDel.NetDel)
				delete(profilePtr.NetDevs, vars.profileDel.NetDel)
			}
			if vars.profileDel.PartDel != "" {
				for diskname, disk := range profilePtr.Disks {
					if _, ok := disk.Partitions[vars.profileDel.PartDel]; ok {
						wwlog.Verbose("Profile: %s, on disk %s, deleting partition: %s", profileId, diskname, vars.profileDel.PartDel)
						delete(disk.Partitions, vars.profileDel.PartDel)
					} else {
						return fmt.Errorf("partition doesn't exist: %s", vars.profileDel.PartDel)
					}
				}
			}
			if vars.profileDel.DiskDel != "" {
				if _, ok := profilePtr.Disks[vars.profileDel.DiskDel]; ok {
					wwlog.Verbose("Profile: %s, deleting disk: %s", profileId, vars.profileDel.DiskDel)
					delete(profilePtr.Disks, vars.profileDel.DiskDel)
				} else {
					return fmt.Errorf("disk doesn't exist: %s", vars.profileDel.DiskDel)
				}
			}
			if vars.profileDel.FsDel != "" {
				if _, ok := profilePtr.FileSystems[vars.profileDel.FsDel]; ok {
					wwlog.Verbose("Profile: %s, deleting filesystem: %s", profileId, vars.profileDel.FsDel)
					delete(profilePtr.FileSystems, vars.profileDel.FsDel)
				} else {
					return fmt.Errorf("filesystem doesn't exist: %s", vars.profileDel.FsDel)
				}
			}
			for _, key := range vars.profileDel.TagsDel {
				delete(profilePtr.Tags, key)
			}
			for key, val := range vars.profileAdd.TagsAdd {
				if profilePtr.Tags == nil {
					profilePtr.Tags = make(map[string]string)
				}
				profilePtr.Tags[key] = val
			}
			for key, val := range vars.profileAdd.IpmiTagsAdd {
				if profilePtr.Ipmi.Tags == nil {
					profilePtr.Ipmi.Tags = make(map[string]string)
				}
				profilePtr.Ipmi.Tags[key] = val
			}
			for _, key := range vars.profileDel.IpmiTagsDel {
				delete(profilePtr.Ipmi.Tags, key)
			}
			if netDev, ok := profilePtr.NetDevs[vars.profileAdd.Net]; ok {
				for _, key := range vars.profileDel.NetTagsDel {
					delete(netDev.Tags, key)
				}
				// Note: original API code used set.TagAdd here (likely a bug),
				// preserving existing behavior by using TagsAdd
				if len(vars.profileAdd.TagsAdd) > 0 && netDev.Tags == nil {
					netDev.Tags = make(map[string]string)
				}
				for key, val := range vars.profileAdd.TagsAdd {
					netDev.Tags[key] = val
				}
			}
			count++
		}

		if !vars.setYes {
			yes := util.Confirm(fmt.Sprintf("Are you sure you want to modify %d profile(s)", count))
			if !yes {
				return nil
			}
		}

		if err := nodeDB.Persist(); err != nil {
			return err
		}
		return warewulfd.DaemonReload()
	}
}
