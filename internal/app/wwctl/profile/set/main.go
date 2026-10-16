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
		if vars.profileAdd.PartName != "" && vars.profileAdd.DiskName == "" {
			return fmt.Errorf("partition and disk must be specified")
		}
		delete(vars.profileConf.Disks, "UNDEF")

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
		profileChanges := map[string][]node.Change{}
		for _, profileId := range args {
			wwlog.Verbose("evaluating profile: %s", profileId)
			profilePtr, err := nodeDB.GetProfilePtr(profileId)
			if err != nil {
				wwlog.Warn("invalid profile: %s", profileId)
				continue
			}
			before := profilePtr.Clone()
			before.Flatten()
			profilePtr.UpdateFrom(&vars.profileConf, changed)
			if vars.profileDel.NetDel != "" {
				if _, ok := profilePtr.NetDevs[vars.profileDel.NetDel]; !ok {
					return fmt.Errorf("network device name doesn't exist: %s", vars.profileDel.NetDel)
				}
				wwlog.Verbose("Profile: %s, Deleting network device: %s", profileId, vars.profileDel.NetDel)
				delete(profilePtr.NetDevs, vars.profileDel.NetDel)
			}
			if vars.profileDel.PartDel != "" {
				if vars.profileAdd.DiskName != "" {
					disk, ok := profilePtr.Disks[vars.profileAdd.DiskName]
					if !ok || disk == nil {
						return fmt.Errorf("disk doesn't exist: %s", vars.profileAdd.DiskName)
					}
					if _, ok := disk.Partitions[vars.profileDel.PartDel]; !ok {
						return fmt.Errorf("partition doesn't exist: %s", vars.profileDel.PartDel)
					}
					wwlog.Verbose("Profile: %s, on disk %s, deleting partition: %s", profileId, vars.profileAdd.DiskName, vars.profileDel.PartDel)
					delete(disk.Partitions, vars.profileDel.PartDel)
				} else {
					found := false
					for diskname, disk := range profilePtr.Disks {
						if _, ok := disk.Partitions[vars.profileDel.PartDel]; ok {
							wwlog.Verbose("Profile: %s, on disk %s, deleting partition: %s", profileId, diskname, vars.profileDel.PartDel)
							delete(disk.Partitions, vars.profileDel.PartDel)
							found = true
						}
					}
					if !found {
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
			if len(vars.profileDel.NetTagsDel) > 0 && vars.profileDel.NetDel != vars.profileAdd.Net {
				if _, ok := profilePtr.NetDevs[vars.profileAdd.Net]; !ok {
					wwlog.Warn("Profile: %s, network device %s is not defined on this profile; no network tags deleted", profileId, vars.profileAdd.Net)
				}
			}
			profilePtr.ApplyTagChanges(vars.profileAdd, vars.profileDel)
			profilePtr.Flatten()
			if before != nil {
				if ch := node.Diff(before, profilePtr); len(ch) > 0 {
					profileChanges[profileId] = ch
				}
			}
		}

		summary := node.FormatChanges(profileChanges)
		if !vars.setYes {
			if summary == "" {
				wwlog.Info("No changes to apply.")
				return nil
			}
			wwlog.Output("%s", summary)
			if !util.Confirm(fmt.Sprintf("Apply these changes to %d profile(s)?", len(profileChanges))) {
				wwlog.Info("No changes made!")
				return nil
			}
		} else {
			wwlog.Output("Applying following changes:\n %s", summary)
		}

		if err := nodeDB.Persist(); err != nil {
			return err
		}
		return warewulfd.DaemonReload()
	}
}
