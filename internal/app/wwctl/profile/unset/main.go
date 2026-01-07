package unset

import (
	"fmt"

	"github.com/spf13/cobra"
	wwctlflags "github.com/warewulf/warewulf/internal/app/wwctl/flags"
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

		// Validate scoping: sub-entity fields require their parent scope flags
		if err := wwctlflags.ValidateUnsetScope(vars.unsetFields, vars.diskname, vars.partname, vars.fsname); err != nil {
			return err
		}

		// Confirmation prompt
		if !vars.unsetYes {
			count := 0
			for _, profileName := range args {
				if _, ok := nodeDB.NodeProfiles[profileName]; ok {
					count++
				}
			}
			if count == 0 {
				return fmt.Errorf("no valid profiles found")
			}
			yes := util.Confirm(fmt.Sprintf("Are you sure you want to modify %d profile(s)", count))
			if !yes {
				return nil
			}
		}

		// Modify profiles directly
		modifiedCount := 0
		for _, profileName := range args {
			profilePtr, ok := nodeDB.NodeProfiles[profileName]
			if !ok {
				wwlog.Warn("invalid profile: %s", profileName)
				if !vars.unsetForce {
					return fmt.Errorf("profile not found: %s", profileName)
				}
				continue
			}

			// Build scope profile: zero-value with map entries for the keys to target
			scopeProfile := node.NewProfile("")
			scopeProfile.NetDevs[vars.netname] = &node.NetDev{}
			if vars.diskname != "" {
				disk := &node.Disk{}
				if vars.partname != "" {
					disk.Partitions = map[string]*node.Partition{vars.partname: {}}
				}
				scopeProfile.Disks = map[string]*node.Disk{vars.diskname: disk}
			}
			if vars.fsname != "" {
				scopeProfile.FileSystems = map[string]*node.FileSystem{vars.fsname: {}}
			}
			changed := func(lopt string) bool {
				boolPtr, ok := vars.unsetFields[lopt]
				return ok && boolPtr != nil && *boolPtr
			}
			profilePtr.UpdateFrom(&scopeProfile, changed)

			// Delete specified tags
			for _, key := range vars.tags {
				delete(profilePtr.Tags, key)
			}
			if profilePtr.Ipmi != nil {
				for _, key := range vars.ipmiTags {
					delete(profilePtr.Ipmi.Tags, key)
				}
			}
			if len(vars.netTags) > 0 {
				if netDev, ok := profilePtr.NetDevs[vars.netname]; ok && netDev != nil {
					for _, key := range vars.netTags {
						delete(netDev.Tags, key)
					}
				}
			}

			// Delete entire objects by name
			for _, name := range vars.netDel {
				delete(profilePtr.NetDevs, name)
			}
			for _, name := range vars.diskDel {
				delete(profilePtr.Disks, name)
			}
			for _, name := range vars.partDel {
				for _, disk := range profilePtr.Disks {
					if disk == nil {
						continue
					}
					delete(disk.Partitions, name)
				}
			}
			for _, name := range vars.fsDel {
				delete(profilePtr.FileSystems, name)
			}

			// Clean up empty structs
			profilePtr.Flatten()

			// Remove any empty map entries left after flattening
			for netName, netDev := range profilePtr.NetDevs {
				if node.ObjectIsEmpty(netDev) {
					delete(profilePtr.NetDevs, netName)
				}
			}
			for diskName, disk := range profilePtr.Disks {
				if node.ObjectIsEmpty(disk) {
					delete(profilePtr.Disks, diskName)
				}
			}
			for fsName, fs := range profilePtr.FileSystems {
				if node.ObjectIsEmpty(fs) {
					delete(profilePtr.FileSystems, fsName)
				}
			}

			modifiedCount++
		}

		if modifiedCount == 0 {
			return fmt.Errorf("no profiles were modified")
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

		wwlog.Info("Successfully unset fields on %d profile(s)", modifiedCount)
		return nil
	}
}
