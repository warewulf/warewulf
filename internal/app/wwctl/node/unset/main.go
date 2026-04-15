package unset

import (
	"fmt"

	"github.com/spf13/cobra"
	wwctlflags "github.com/warewulf/warewulf/internal/app/wwctl/flags"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Check if any fields were specified
		anyFieldSet := false
		for _, boolPtr := range vars.unsetFields {
			if *boolPtr {
				anyFieldSet = true
				break
			}
		}
		anyFieldSet = anyFieldSet || len(vars.tags) > 0 || len(vars.ipmiTags) > 0 || len(vars.netTags) > 0 ||
			len(vars.netDel) > 0 || len(vars.diskDel) > 0 || len(vars.partDel) > 0 || len(vars.fsDel) > 0
		if !anyFieldSet {
			return fmt.Errorf("no fields specified to unset")
		}

		nodeDB, err := node.New()
		if err != nil {
			return fmt.Errorf("failed to load node database: %w", err)
		}

		// Expand hostlist patterns
		args = hostlist.Expand(args)

		// Validate scoping: sub-entity fields require their parent scope flags
		if err := wwctlflags.ValidateUnsetScope(vars.unsetFields, vars.diskname, vars.partname, vars.fsname); err != nil {
			return err
		}

		// Confirmation prompt
		if !vars.unsetYes {
			count := 0
			for _, nodeName := range args {
				if _, ok := nodeDB.Nodes[nodeName]; ok {
					count++
				}
			}
			if count == 0 {
				return fmt.Errorf("no valid nodes found")
			}
			yes := util.Confirm(fmt.Sprintf("Are you sure you want to modify %d node(s)", count))
			if !yes {
				return nil
			}
		}

		// Modify nodes directly
		modifiedCount := 0
		for _, nodeName := range args {
			nodePtr, ok := nodeDB.Nodes[nodeName]
			if !ok {
				wwlog.Warn("invalid node: %s", nodeName)
				if !vars.unsetForce {
					return fmt.Errorf("node not found: %s", nodeName)
				}
				continue
			}

			// Build scope node: zero-value with map entries for the keys to target
			scopeNode := node.NewNode("")
			scopeNode.NetDevs[vars.netname] = &node.NetDev{}
			if vars.diskname != "" {
				disk := &node.Disk{}
				if vars.partname != "" {
					disk.Partitions = map[string]*node.Partition{vars.partname: {}}
				}
				scopeNode.Disks = map[string]*node.Disk{vars.diskname: disk}
			}
			if vars.fsname != "" {
				scopeNode.FileSystems = map[string]*node.FileSystem{vars.fsname: {}}
			}
			changed := func(lopt string) bool {
				boolPtr, ok := vars.unsetFields[lopt]
				return ok && boolPtr != nil && *boolPtr
			}
			nodePtr.UpdateFrom(&scopeNode, changed)

			// Delete specified tags
			for _, key := range vars.tags {
				delete(nodePtr.Tags, key)
			}
			if nodePtr.Ipmi != nil {
				for _, key := range vars.ipmiTags {
					delete(nodePtr.Ipmi.Tags, key)
				}
			}
			if len(vars.netTags) > 0 {
				if netDev, ok := nodePtr.NetDevs[vars.netname]; ok && netDev != nil {
					for _, key := range vars.netTags {
						delete(netDev.Tags, key)
					}
				}
			}

			// Delete entire objects by name
			for _, name := range vars.netDel {
				delete(nodePtr.NetDevs, name)
			}
			for _, name := range vars.diskDel {
				delete(nodePtr.Disks, name)
			}
			for _, name := range vars.partDel {
				if vars.diskname != "" {
					disk, ok := nodePtr.Disks[vars.diskname]
					if !ok || disk == nil {
						return fmt.Errorf("disk doesn't exist: %s", vars.diskname)
					}
					if _, ok := disk.Partitions[name]; !ok {
						return fmt.Errorf("partition doesn't exist: %s", name)
					}
					delete(disk.Partitions, name)
				} else {
					found := false
					for _, disk := range nodePtr.Disks {
						if disk == nil {
							continue
						}
						if _, ok := disk.Partitions[name]; ok {
							delete(disk.Partitions, name)
							found = true
						}
					}
					if !found {
						return fmt.Errorf("partition doesn't exist: %s", name)
					}
				}
			}
			for _, name := range vars.fsDel {
				delete(nodePtr.FileSystems, name)
			}

			// Clean up empty structs
			nodePtr.Flatten()

			// Remove any empty map entries left after flattening
			for netName, netDev := range nodePtr.NetDevs {
				if node.ObjectIsEmpty(netDev) {
					delete(nodePtr.NetDevs, netName)
				}
			}
			for diskName, disk := range nodePtr.Disks {
				if node.ObjectIsEmpty(disk) {
					delete(nodePtr.Disks, diskName)
				}
			}
			for fsName, fs := range nodePtr.FileSystems {
				if node.ObjectIsEmpty(fs) {
					delete(nodePtr.FileSystems, fsName)
				}
			}

			modifiedCount++
		}

		if modifiedCount == 0 {
			return fmt.Errorf("no nodes were modified")
		}

		// Save changes
		if err := nodeDB.Persist(); err != nil {
			return fmt.Errorf("failed to persist changes: %w", err)
		}

		// Reload daemon
		if err := warewulfd.DaemonReload(); err != nil {
			wwlog.Warn("failed to reload daemon: %v", err)
			// Don't fail - changes were saved
		}

		wwlog.Info("Successfully unset fields on %d node(s)", modifiedCount)
		return nil
	}
}
