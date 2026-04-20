package unset

import (
	"fmt"

	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// ValidateScopeRequirements checks that sub-entity unset flags have the
// required scoping flags, using the scope map produced by CreateUnsetFlags.
// Scope values: "disk" requires --diskname; "disk,part" requires both
// --diskname and --partname; "fs" requires --fsname.
func ValidateScopeRequirements(vars *Vars) error {
	for flagName, boolPtr := range vars.UnsetFields {
		if boolPtr == nil || !*boolPtr {
			continue
		}
		switch vars.UnsetScopes[flagName] {
		case "disk,part":
			if vars.Diskname == "" || vars.Partname == "" {
				return fmt.Errorf("--diskname and --partname must be specified with --%s", flagName)
			}
		case "disk":
			if vars.Diskname == "" {
				return fmt.Errorf("--diskname must be specified with --%s", flagName)
			}
		case "fs":
			if vars.Fsname == "" {
				return fmt.Errorf("--fsname must be specified with --%s", flagName)
			}
		}
	}
	return nil
}

// WarnDeletions prints a warning line for each sub-entity that will be deleted
// entirely (scoping flag given with no corresponding sub-field flags). Call
// before prompting for confirmation so the user knows what will be removed.
func WarnDeletions(vars *Vars) {
	if vars.NetnameChanged && !HasScopedFieldSet(vars, "net") && len(vars.NetTags) == 0 {
		wwlog.Warn("network device %q will be removed entirely", vars.Netname)
	}
	if vars.Diskname != "" && vars.Partname == "" && !HasScopedFieldSet(vars, "disk") {
		wwlog.Warn("disk %q will be removed entirely", vars.Diskname)
	}
	if vars.Fsname != "" && !HasScopedFieldSet(vars, "fs") {
		wwlog.Warn("filesystem %q will be removed entirely", vars.Fsname)
	}
	if vars.Partname != "" && !HasScopedFieldSet(vars, "disk,part") {
		if vars.Diskname != "" {
			wwlog.Warn("partition %q on disk %q will be removed entirely", vars.Partname, vars.Diskname)
		} else {
			wwlog.Warn("partition %q will be removed from all disks", vars.Partname)
		}
	}
}

// HasScopedFieldSet returns true if any field in vars.UnsetFields is set and
// has exactly the given scope value in vars.UnsetScopes.
func HasScopedFieldSet(vars *Vars, scope string) bool {
	for flagName, boolPtr := range vars.UnsetFields {
		if boolPtr != nil && *boolPtr && vars.UnsetScopes[flagName] == scope {
			return true
		}
	}
	return false
}

// Entity is satisfied by *node.Node and *node.Profile.
type Entity interface {
	GetProfile() *node.Profile
	UpdateFromProfile(src *node.Profile, changed func(string) bool)
	Flatten()
}

// UpdateEntity applies all unset operations to entity, which must be a
// *node.Node or *node.Profile.
func UpdateEntity(entity Entity, vars *Vars) error {
	target := entity.GetProfile()
	// Build scope: a zero-valued Profile with map entries for the targeted
	// sub-entity keys so UpdateFrom knows which sub-entity to clear fields on.
	// Only add sub-entities that actually have fields being unset — adding an
	// entry unconditionally causes recursiveUpdateFrom to create stubs in the
	// target, which defeats existence checks in the deletion logic below.
	scope := node.NewProfile("")
	if HasScopedFieldSet(vars, "net") {
		scope.NetDevs[vars.Netname] = &node.NetDev{}
	}
	if vars.Diskname != "" {
		hasDiskFields := HasScopedFieldSet(vars, "disk")
		hasPartFields := HasScopedFieldSet(vars, "disk,part")
		if hasDiskFields || hasPartFields {
			disk := &node.Disk{}
			if vars.Partname != "" && hasPartFields {
				disk.Partitions = map[string]*node.Partition{vars.Partname: {}}
			}
			scope.Disks = map[string]*node.Disk{vars.Diskname: disk}
		}
	}
	if vars.Fsname != "" && HasScopedFieldSet(vars, "fs") {
		scope.FileSystems = map[string]*node.FileSystem{vars.Fsname: {}}
	}

	changed := func(lopt string) bool {
		boolPtr, ok := vars.UnsetFields[lopt]
		return ok && boolPtr != nil && *boolPtr
	}
	entity.UpdateFromProfile(&scope, changed)

	// Delete specified tags
	for _, key := range vars.Tags {
		delete(target.Tags, key)
	}
	if target.Ipmi != nil {
		for _, key := range vars.IpmiTags {
			delete(target.Ipmi.Tags, key)
		}
	}
	if len(vars.NetTags) > 0 {
		if netDev, ok := target.NetDevs[vars.Netname]; ok && netDev != nil {
			for _, key := range vars.NetTags {
				delete(netDev.Tags, key)
			}
		}
	}

	// Delete entire sub-entities when scoping flag given with no sub-fields.
	// Net: also guard against --nettag (tag ops use --netname for scoping).
	// Disk: also guard against --partname (that scopes a partition operation).
	if vars.NetnameChanged && !HasScopedFieldSet(vars, "net") && len(vars.NetTags) == 0 {
		delete(target.NetDevs, vars.Netname)
	}
	if vars.Diskname != "" && vars.Partname == "" && !HasScopedFieldSet(vars, "disk") {
		delete(target.Disks, vars.Diskname)
	}
	if vars.Fsname != "" && !HasScopedFieldSet(vars, "fs") {
		delete(target.FileSystems, vars.Fsname)
	}
	if vars.Partname != "" && !HasScopedFieldSet(vars, "disk,part") {
		if vars.Diskname != "" {
			disk, ok := target.Disks[vars.Diskname]
			if !ok || disk == nil {
				return fmt.Errorf("disk doesn't exist: %s", vars.Diskname)
			}
			if _, ok := disk.Partitions[vars.Partname]; !ok {
				return fmt.Errorf("partition doesn't exist: %s", vars.Partname)
			}
			delete(disk.Partitions, vars.Partname)
		} else {
			found := false
			for _, disk := range target.Disks {
				if disk == nil {
					continue
				}
				if _, ok := disk.Partitions[vars.Partname]; ok {
					delete(disk.Partitions, vars.Partname)
					found = true
				}
			}
			if !found {
				return fmt.Errorf("partition doesn't exist: %s", vars.Partname)
			}
		}
	}

	entity.Flatten()

	// Remove any empty map entries left after flattening
	for netName, netDev := range target.NetDevs {
		if node.ObjectIsEmpty(netDev) {
			delete(target.NetDevs, netName)
		}
	}
	for diskName, disk := range target.Disks {
		if node.ObjectIsEmpty(disk) {
			delete(target.Disks, diskName)
		}
	}
	for fsName, fs := range target.FileSystems {
		if node.ObjectIsEmpty(fs) {
			delete(target.FileSystems, fsName)
		}
	}

	return nil
}
