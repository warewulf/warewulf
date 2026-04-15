package unset

import (
	"fmt"

	"github.com/warewulf/warewulf/internal/pkg/node"
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
	scope := node.NewProfile("")
	scope.NetDevs[vars.Netname] = &node.NetDev{}
	if vars.Diskname != "" {
		disk := &node.Disk{}
		if vars.Partname != "" {
			disk.Partitions = map[string]*node.Partition{vars.Partname: {}}
		}
		scope.Disks = map[string]*node.Disk{vars.Diskname: disk}
	}
	if vars.Fsname != "" {
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

	// Delete entire objects by name
	for _, name := range vars.NetDel {
		delete(target.NetDevs, name)
	}
	for _, name := range vars.DiskDel {
		delete(target.Disks, name)
	}
	for _, name := range vars.PartDel {
		if vars.Diskname != "" {
			disk, ok := target.Disks[vars.Diskname]
			if !ok || disk == nil {
				return fmt.Errorf("disk doesn't exist: %s", vars.Diskname)
			}
			if _, ok := disk.Partitions[name]; !ok {
				return fmt.Errorf("partition doesn't exist: %s", name)
			}
			delete(disk.Partitions, name)
		} else {
			found := false
			for _, disk := range target.Disks {
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
	for _, name := range vars.FsDel {
		delete(target.FileSystems, name)
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
