package apiprofile

import (
	"fmt"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// NodeSet is the wwapiv1 implmentation for updating nodeinfo fields.
func ProfileSet(set *wwapiv1.ConfSetParameter) (err error) {
	if set == nil {
		return fmt.Errorf("ProfileAddParameter is nil")
	}
	nodeDB, _, err := ProfileSetParameterCheck(set)
	if err != nil {
		return errors.Wrap(err, "profile set parameters are wrong")
	}
	if err = nodeDB.Persist(); err != nil {
		return err
	}
	if err = warewulfd.DaemonReload(); err != nil {
		return err
	}
	return
}

/*
NodeSetParameterCheck does error checking and returns a modified
NodeYml which than can be persisted
*/
func ProfileSetParameterCheck(set *wwapiv1.ConfSetParameter) (nodeDB node.NodeYaml, count uint, err error) {
	nodeDB, err = node.New()
	if err != nil {
		wwlog.Error("Could not open configuration: %s", err)
		return
	}
	//func AbstractSetParameterCheck(set *wwapiv1.ConfSetParameter, confMap map[string]*node.NodeConf, confs []string) (count uint, err error) {
	if set == nil {
		err = fmt.Errorf("profile set parameter is nil")
		return
	}
	if set.ConfList == nil {
		err = fmt.Errorf("node nodes to set!")
		return
	}
	confs := nodeDB.ListAllProfiles()
	// Note: This does not do expansion on the nodes.
	if set.AllConfs || (len(set.ConfList) == 0) {
		wwlog.Warn("this command will modify all nodes/profiles")
	} else if len(confs) == 0 {
		wwlog.Warn("no nodes/profiles found")
		return
	} else {
		confs = set.ConfList
	}
	//var confobject node.NodeConf
	for _, p := range set.ConfList {
		if util.InSlice(set.ConfList, p) {
			wwlog.Verbose("evaluating profile: %s", p)
			if _, ok := nodeDB.NodeProfiles[p]; !ok {
				continue
			}
			err = yaml.Unmarshal([]byte(set.NodeConfYaml), nodeDB.NodeProfiles[p])
			if err != nil {
				return
			}
			if set.NetdevDelete != "" {
				if _, ok := nodeDB.NodeProfiles[p].NetDevs[set.NetdevDelete]; !ok {
					err = fmt.Errorf("network device name doesn't exist: %s", set.NetdevDelete)
					wwlog.Error(fmt.Sprintf("%v", err.Error()))
					return
				}
				wwlog.Verbose("Profile: %s, Deleting network device: %s", p, set.NetdevDelete)
				delete(nodeDB.NodeProfiles[p].NetDevs, set.NetdevDelete)
			}
			if set.PartitionDelete != "" {
				for diskname, disk := range nodeDB.NodeProfiles[p].Disks {
					if _, ok := disk.Partitions[set.PartitionDelete]; ok {
						wwlog.Verbose("Node: %s, on disk %, deleting partition: %s", p, diskname, set.PartitionDelete)
						delete(disk.Partitions, set.PartitionDelete)
					} else {
						return nodeDB, count, fmt.Errorf("partition doesn't exist: %s", set.PartitionDelete)

					}
				}
			}
			if set.DiskDelete != "" {
				if _, ok := nodeDB.NodeProfiles[p].Disks[set.DiskDelete]; ok {
					wwlog.Verbose("Node: %s, deleting disk: %s", p, set.DiskDelete)
					delete(nodeDB.NodeProfiles[p].Disks, set.DiskDelete)
				} else {
					return nodeDB, count, fmt.Errorf("disk doesn't exist: %s", set.DiskDelete)
				}
			}
			if set.FilesystemDelete != "" {
				if _, ok := nodeDB.NodeProfiles[p].FileSystems[set.FilesystemDelete]; ok {
					wwlog.Verbose("Node: %s, deleting filesystem: %s", p, set.FilesystemDelete)
					delete(nodeDB.NodeProfiles[p].FileSystems, set.FilesystemDelete)
				} else {
					return nodeDB, count, fmt.Errorf("disk doesn't exist: %s", set.FilesystemDelete)
				}
			}
			for _, key := range set.TagDel {
				delete(nodeDB.NodeProfiles[p].Tags, key)
			}
			for key, val := range set.TagAdd {
				if nodeDB.NodeProfiles[p].Tags == nil {
					nodeDB.NodeProfiles[p].Tags = make(map[string]string)
				}
				nodeDB.NodeProfiles[p].Tags[key] = val
			}
			for key, val := range set.IpmiTagAdd {
				if nodeDB.NodeProfiles[p].Ipmi.Tags == nil {
					nodeDB.NodeProfiles[p].Ipmi.Tags = make(map[string]string)
				}
				nodeDB.NodeProfiles[p].Ipmi.Tags[key] = val
			}
			for _, key := range set.IpmiTagDel {
				delete(nodeDB.NodeProfiles[p].Ipmi.Tags, key)
			}
			if _, ok := nodeDB.NodeProfiles[p].NetDevs[set.Netdev]; ok {
				for _, key := range set.NetTagDel {
					delete(nodeDB.NodeProfiles[p].NetDevs[set.Netdev].Tags, key)
				}
				for key, val := range set.TagAdd {
					nodeDB.NodeProfiles[p].NetDevs[set.Netdev].Tags[key] = val
				}
			}
			count++
		}
	}
	return
}
