package apinode

import (
	"fmt"

	"dario.cat/mergo"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd/daemon"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
	"gopkg.in/yaml.v3"
)

// NodeSet is the wwapiv1 implmentation for updating node fields.
func NodeSet(set *wwapiv1.ConfSetParameter) (err error) {
	if set == nil {
		return fmt.Errorf("NodeSetParameter is nil")
	}
	var nodeDB node.NodesYaml
	nodeDB, _, err = NodeSetParameterCheck(set)
	if err != nil {
		return err
	}
	if err = nodeDB.Persist(); err != nil {
		return err
	}
	if err = daemon.DaemonReload(); err != nil {
		return err
	}
	return
}

/*
NodeSetParameterCheck does error checking and returns a modified
NodeYml which than can be persisted
*/
func NodeSetParameterCheck(set *wwapiv1.ConfSetParameter) (nodeDB node.NodesYaml, count uint, err error) {
	nodeDB, err = node.New()
	if err != nil {
		wwlog.Error("Could not open configuration: %s", err)
		return
	}
	if set == nil {
		err = fmt.Errorf("node set parameter is nil")
		return
	}
	if set.ConfList == nil {
		err = fmt.Errorf("node nodes to set")
		return
	}
	confs := nodeDB.ListAllNodes()
	// Note: This does not do expansion on the nodes.
	if set.AllConfs || (len(set.ConfList) == 0) {
		wwlog.Warn("this command will modify all nodes/profiles")
	} else if len(confs) == 0 {
		wwlog.Warn("no nodes/profiles found")
		return
	}
	for _, nId := range set.ConfList {
		if util.InSlice(set.ConfList, nId) {
			wwlog.Debug("evaluating node: %s", nId)
			var nodePtr *node.Node
			nodePtr, err = nodeDB.GetNodeOnlyPtr(nId)
			if err != nil {
				wwlog.Warn("invalid node: %s", nId)
				continue
			}
			newConf := node.EmptyNode()
			err = yaml.Unmarshal([]byte(set.NodeConfYaml), &newConf)
			if err != nil {
				return
			}
			// merge in
			err = mergo.Merge(nodePtr, &newConf, mergo.WithOverride)
			if err != nil {
				return
			}
			if set.NetdevDelete != "" {
				if _, ok := nodePtr.NetDevs[set.NetdevDelete]; !ok {
					err = fmt.Errorf("network device name doesn't exist: %s", set.NetdevDelete)
					wwlog.Error(fmt.Sprintf("%v", err.Error()))
					return
				}
				wwlog.Verbose("Profile: %s, Deleting network device: %s", nId, set.NetdevDelete)
				delete(nodePtr.NetDevs, set.NetdevDelete)
			}
			if set.PartitionDelete != "" {
				for diskname, disk := range nodePtr.Disks {
					if _, ok := disk.Partitions[set.PartitionDelete]; ok {
						wwlog.Verbose("Node: %s, on disk %, deleting partition: %s", nId, diskname, set.PartitionDelete)
						delete(disk.Partitions, set.PartitionDelete)
					} else {
						return nodeDB, count, fmt.Errorf("partition doesn't exist: %s", set.PartitionDelete)

					}
				}
			}
			if set.DiskDelete != "" {
				if _, ok := nodePtr.Disks[set.DiskDelete]; ok {
					wwlog.Verbose("Node: %s, deleting disk: %s", nId, set.DiskDelete)
					delete(nodePtr.Disks, set.DiskDelete)
				} else {
					return nodeDB, count, fmt.Errorf("disk doesn't exist: %s", set.DiskDelete)
				}
			}
			if set.FilesystemDelete != "" {
				if _, ok := nodePtr.FileSystems[set.FilesystemDelete]; ok {
					wwlog.Verbose("Node: %s, deleting filesystem: %s", nId, set.FilesystemDelete)
					delete(nodePtr.FileSystems, set.FilesystemDelete)
				} else {
					return nodeDB, count, fmt.Errorf("disk doesn't exist: %s", set.FilesystemDelete)
				}
			}
			for _, key := range set.TagDel {
				delete(nodePtr.Tags, key)
			}
			for key, val := range set.TagAdd {
				if nodePtr.Tags == nil {
					nodePtr.Tags = make(map[string]string)
				}
				nodePtr.Tags[key] = val
			}
			for key, val := range set.IpmiTagAdd {
				if nodePtr.Ipmi.Tags == nil {
					nodePtr.Ipmi.Tags = make(map[string]string)
				}
				nodePtr.Ipmi.Tags[key] = val
			}
			for _, key := range set.IpmiTagDel {
				delete(nodePtr.Ipmi.Tags, key)
			}
			if _, ok := nodePtr.NetDevs[set.Netdev]; ok {
				for _, key := range set.NetTagDel {
					delete(nodePtr.NetDevs[set.Netdev].Tags, key)
				}
				for key, val := range set.NetTagAdd {
					nodePtr.NetDevs[set.Netdev].Tags[key] = val
				}
			}
			count++
		}
	}
	return
}
