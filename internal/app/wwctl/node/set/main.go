package set

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) error {
		// remove the default network as the all network values are assigned
		// to this network
		if !node.ObjectIsEmpty(vars.nodeConf.NetDevs["UNDEF"]) || len(vars.nodeAdd.NetTagsAdd) > 0 {
			netDev := *vars.nodeConf.NetDevs["UNDEF"]
			vars.nodeConf.NetDevs[vars.nodeAdd.Net] = &netDev
			vars.nodeConf.NetDevs[vars.nodeAdd.Net].Tags = vars.nodeAdd.NetTagsAdd
		}
		delete(vars.nodeConf.NetDevs, "UNDEF")
		if vars.nodeAdd.FsName != "" {
			if !strings.HasPrefix(vars.nodeAdd.FsName, "/dev") {
				if vars.nodeAdd.FsName == vars.nodeAdd.PartName {
					vars.nodeAdd.FsName = "/dev/disk/by-partlabel/" + vars.nodeAdd.PartName
				} else {
					return fmt.Errorf("filesystems need to have a underlying blockdev")
				}
			}
			fs := *vars.nodeConf.FileSystems["UNDEF"]
			vars.nodeConf.FileSystems[vars.nodeAdd.FsName] = &fs
		}
		delete(vars.nodeConf.FileSystems, "UNDEF")
		if vars.nodeAdd.DiskName != "" && vars.nodeAdd.PartName != "" {
			prt := *vars.nodeConf.Disks["UNDEF"].Partitions["UNDEF"]
			vars.nodeConf.Disks["UNDEF"].Partitions[vars.nodeAdd.PartName] = &prt
			delete(vars.nodeConf.Disks["UNDEF"].Partitions, "UNDEF")
			dsk := *vars.nodeConf.Disks["UNDEF"]
			vars.nodeConf.Disks[vars.nodeAdd.DiskName] = &dsk
		}
		if vars.nodeAdd.PartName != "" && vars.nodeAdd.DiskName == "" {
			return fmt.Errorf("partition and disk must be specified")
		}
		delete(vars.nodeConf.Disks, "UNDEF")
		vars.nodeConf.Ipmi.Tags = vars.nodeAdd.IpmiTagsAdd

		nodeDB, err := node.New()
		if err != nil {
			return fmt.Errorf("could not open configuration: %w", err)
		}

		args = hostlist.Expand(args)
		if len(args) == 0 && !vars.setNodeAll {
			return fmt.Errorf("no nodes specified; use --all to modify all nodes")
		}
		if vars.setNodeAll {
			args = nodeDB.ListAllNodes()
			wwlog.Warn("this command will modify all nodes")
		} else if len(nodeDB.ListAllNodes()) == 0 {
			wwlog.Warn("no nodes/profiles found")
			return nil
		}

		changed := cmd.Flags().Changed
		var count uint
		for _, nId := range args {
			wwlog.Debug("evaluating node: %s", nId)
			nodePtr, err := nodeDB.GetNodeOnlyPtr(nId)
			if err != nil {
				wwlog.Warn("invalid node: %s", nId)
				continue
			}
			nodePtr.UpdateFrom(&vars.nodeConf, changed)
			if vars.nodeDel.NetDel != "" {
				if _, ok := nodePtr.NetDevs[vars.nodeDel.NetDel]; !ok {
					return fmt.Errorf("network device name doesn't exist: %s", vars.nodeDel.NetDel)
				}
				wwlog.Verbose("Node: %s, Deleting network device: %s", nId, vars.nodeDel.NetDel)
				delete(nodePtr.NetDevs, vars.nodeDel.NetDel)
			}
			if vars.nodeDel.PartDel != "" {
				if vars.nodeAdd.DiskName != "" {
					disk, ok := nodePtr.Disks[vars.nodeAdd.DiskName]
					if !ok || disk == nil {
						return fmt.Errorf("disk doesn't exist: %s", vars.nodeAdd.DiskName)
					}
					if _, ok := disk.Partitions[vars.nodeDel.PartDel]; !ok {
						return fmt.Errorf("partition doesn't exist: %s", vars.nodeDel.PartDel)
					}
					wwlog.Verbose("Node: %s, on disk %s, deleting partition: %s", nId, vars.nodeAdd.DiskName, vars.nodeDel.PartDel)
					delete(disk.Partitions, vars.nodeDel.PartDel)
				} else {
					found := false
					for diskname, disk := range nodePtr.Disks {
						if _, ok := disk.Partitions[vars.nodeDel.PartDel]; ok {
							wwlog.Verbose("Node: %s, on disk %s, deleting partition: %s", nId, diskname, vars.nodeDel.PartDel)
							delete(disk.Partitions, vars.nodeDel.PartDel)
							found = true
						}
					}
					if !found {
						return fmt.Errorf("partition doesn't exist: %s", vars.nodeDel.PartDel)
					}
				}
			}
			if vars.nodeDel.DiskDel != "" {
				if _, ok := nodePtr.Disks[vars.nodeDel.DiskDel]; ok {
					wwlog.Verbose("Node: %s, deleting disk: %s", nId, vars.nodeDel.DiskDel)
					delete(nodePtr.Disks, vars.nodeDel.DiskDel)
				} else {
					return fmt.Errorf("disk doesn't exist: %s", vars.nodeDel.DiskDel)
				}
			}
			if vars.nodeDel.FsDel != "" {
				if _, ok := nodePtr.FileSystems[vars.nodeDel.FsDel]; ok {
					wwlog.Verbose("Node: %s, deleting filesystem: %s", nId, vars.nodeDel.FsDel)
					delete(nodePtr.FileSystems, vars.nodeDel.FsDel)
				} else {
					return fmt.Errorf("filesystem doesn't exist: %s", vars.nodeDel.FsDel)
				}
			}
			for _, key := range vars.nodeDel.TagsDel {
				delete(nodePtr.Tags, key)
			}
			for key, val := range vars.nodeAdd.TagsAdd {
				if nodePtr.Tags == nil {
					nodePtr.Tags = make(map[string]string)
				}
				nodePtr.Tags[key] = val
			}
			for key, val := range vars.nodeAdd.IpmiTagsAdd {
				if nodePtr.Ipmi.Tags == nil {
					nodePtr.Ipmi.Tags = make(map[string]string)
				}
				nodePtr.Ipmi.Tags[key] = val
			}
			for _, key := range vars.nodeDel.IpmiTagsDel {
				delete(nodePtr.Ipmi.Tags, key)
			}
			if netDev, ok := nodePtr.NetDevs[vars.nodeAdd.Net]; ok {
				for _, key := range vars.nodeDel.NetTagsDel {
					delete(netDev.Tags, key)
				}
				if len(vars.nodeAdd.NetTagsAdd) > 0 && netDev.Tags == nil {
					netDev.Tags = make(map[string]string)
				}
				for key, val := range vars.nodeAdd.NetTagsAdd {
					netDev.Tags[key] = val
				}
			}
			nodePtr.Flatten()
			count++
		}

		if !vars.setYes {
			if !util.Confirm(fmt.Sprintf("Are you sure you want to modify %d nodes(s)", count)) {
				return nil
			}
		}

		if err := nodeDB.Persist(); err != nil {
			return err
		}
		return warewulfd.DaemonReload()
	}
}
