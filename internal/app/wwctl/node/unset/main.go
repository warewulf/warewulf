package unset

import (
	"fmt"

	"github.com/spf13/cobra"
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

		// Load registry directly (NOT using API layer)
		nodeDB, err := node.New()
		if err != nil {
			return fmt.Errorf("failed to load node database: %w", err)
		}

		// Expand hostlist patterns
		args = hostlist.Expand(args)

		// Build scoping parameters
		scope := node.UnsetScope{
			NetName:  vars.netname,
			DiskName: vars.diskname,
			PartName: vars.partname,
			FsName:   vars.fsname,
		}
		netname := vars.netname

		// Validate scoping: sub-entity fields require their parent scope flags
		if err := node.ValidateUnsetScope(vars.unsetFields, scope); err != nil {
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

			// Zero the flagged fields using optimized single-pass reflection
			node.ApplyUnsetFields(nodePtr, vars.unsetFields, scope)

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
				if netDev, ok := nodePtr.NetDevs[netname]; ok && netDev != nil {
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
				for _, disk := range nodePtr.Disks {
					if disk == nil {
						continue
					}
					delete(disk.Partitions, name)
				}
			}
			for _, name := range vars.fsDel {
				delete(nodePtr.FileSystems, name)
			}

			// Clean up empty structs
			nodePtr.Flatten()

			// Remove any empty network devices left after flattening
			// Flatten() doesn't remove individual map entries, only zeros entire fields
			for netName, netDev := range nodePtr.NetDevs {
				if node.ObjectIsEmpty(netDev) {
					delete(nodePtr.NetDevs, netName)
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
