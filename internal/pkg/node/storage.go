package node

import (
	"errors"
	"sort"
	"strconv"

	types_3_4 "github.com/coreos/ignition/v2/config/v3_2/types"
	"github.com/coreos/vcontext/path"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func (node *Node) DiskList() (disks []*Disk) {
	names := make([]string, 0, len(node.Disks))
	for name := range node.Disks {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		disk := node.Disks[name]
		if disk != nil {
			disks = append(disks, disk)
		}
	}
	return disks
}

func (node *Node) FileSystemList() (fs []*FileSystem) {
	names := make([]string, 0, len(node.FileSystems))
	for name := range node.FileSystems {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fsys := node.FileSystems[name]
		if fsys != nil {
			fs = append(fs, fsys)
		}
	}
	return fs
}

func (disk *Disk) Id() string {
	return disk.id
}

func (disk *Disk) PartitionList() (partitions []*Partition) {
	names := make([]string, 0, len(disk.Partitions))
	for name := range disk.Partitions {
		names = append(names, name)
	}
	sort.SliceStable(names, func(i, j int) bool {
		pi := disk.Partitions[names[i]]
		pj := disk.Partitions[names[j]]

		// Try to compare by Number (as int)
		ni, erri := strconv.Atoi(pi.Number)
		nj, errj := strconv.Atoi(pj.Number)
		if erri == nil && errj == nil {
			if ni != nj {
				return ni < nj
			}
		} else if erri == nil {
			return true // i has a number, j does not
		} else if errj == nil {
			return false // j has a number, i does not
		}

		// Next, compare by id if set
		if pi.id != "" && pj.id != "" {
			if pi.id != pj.id {
				return pi.id < pj.id
			}
		} else if pi.id != "" {
			return true // i has id, j does not
		} else if pj.id != "" {
			return false // j has id, i does not
		}

		// Fallback: compare by name (map key)
		return names[i] < names[j]
	})
	for _, name := range names {
		part := disk.Partitions[name]
		if part != nil {
			partitions = append(partitions, part)
		}
	}
	return partitions
}

func (partition *Partition) Id() string {
	return partition.id
}

func (fs *FileSystem) Id() string {
	return fs.id
}

/*
Create a ignition struct class for ignition
*/
func (node *Node) GetIgnitionStorage() (stor types_3_4.Storage, rep string, err error) {
	var fileSystems []types_3_4.Filesystem
	for fsdevice, fs := range node.FileSystems {
		var mountOptions []types_3_4.MountOption
		for _, opt := range fs.MountOptions {
			mountOptions = append(mountOptions, types_3_4.MountOption(opt))
		}
		var fsOption []types_3_4.FilesystemOption
		for _, opt := range fs.Options {
			fsOption = append(fsOption, types_3_4.FilesystemOption(opt))
		}
		wipe := fs.WipeFileSystem
		myFs := types_3_4.Filesystem{
			Device:         fsdevice,
			Path:           &fs.Path,
			WipeFilesystem: &wipe,
		}
		if fs.Format != "" {
			myFs.Format = &fs.Format
		}
		if fs.Label != "" {
			myFs.Label = &fs.Label
		}
		if fs.MountOptions != "" {
			myFs.MountOptions = mountOptions
		}
		if len(fs.Options) != 0 {
			myFs.Options = fsOption
		}
		if fs.Uuid != "" {
			myFs.UUID = &fs.Uuid
		}
		wwlog.Debug("created file system struct: %v", myFs)
		fileSystems = append(fileSystems, myFs)
	}
	sort.SliceStable(fileSystems, func(i int, j int) bool {
		return fileSystems[i].Device < fileSystems[j].Device
	})
	var disks []types_3_4.Disk
	for diskDev, disk := range node.Disks {
		var partitions []types_3_4.Partition
		for partlabel, part := range disk.Partitions {
			resize := part.Resize
			shouldExist := part.ShouldExist
			wipe := part.WipePartitionEntry
			label := partlabel
			var number int
			if part.Number != "" {
				number, err = strconv.Atoi(part.Number)
				if err != nil {
					return
				}
			}
			myPart := types_3_4.Partition{
				Label:              &label,
				Number:             number,
				ShouldExist:        &shouldExist,
				WipePartitionEntry: &wipe,
			}
			if part.Guid != "" {
				myPart.GUID = &part.Guid
			}
			if part.Resize {
				myPart.Resize = &resize
			}
			if part.SizeMiB != "" {
				var size int
				size, err = strconv.Atoi(part.SizeMiB)
				if err != nil {
					return
				}
				myPart.SizeMiB = &size
			}
			if part.StartMiB != "" {
				var start int
				start, err = strconv.Atoi(part.SizeMiB)
				if err != nil {
					return
				}
				myPart.StartMiB = &start
			}
			if part.TypeGuid != "" {
				myPart.TypeGUID = &part.TypeGuid
			}
			partitions = append(partitions, myPart)
		}
		sort.SliceStable(partitions, func(i int, j int) bool {
			if partitions[i].Number == partitions[j].Number {
				if partitions[i].SizeMiB != nil && partitions[j].SizeMiB == nil {
					return true
				}
				if partitions[j].SizeMiB != nil && partitions[i].SizeMiB == nil {
					return false
				}
				return *partitions[i].SizeMiB < *partitions[j].SizeMiB
			}
			return partitions[i].Number < partitions[j].Number
		})
		wipe := disk.WipeTable
		disks = append(disks, types_3_4.Disk{
			Device:     diskDev,
			Partitions: partitions,
			WipeTable:  &wipe,
		})
	}
	stor = types_3_4.Storage{
		Disks:       disks,
		Filesystems: fileSystems,
	}
	report := stor.Validate(path.ContextPath{})
	if report.IsFatal() {
		err = errors.New(report.String())
	}
	rep = report.String()
	return
}

type MyIgnition struct {
	Version string `json:"version"`
}
type SimpleIgnitionConfig struct {
	Ignition MyIgnition        `json:"ignition"`
	Storage  types_3_4.Storage `json:"storage"`
}

/*
Get a simple config which can be marshalled to json
*/
func (node *Node) GetIgnitionConfig() (conf SimpleIgnitionConfig, rep string, err error) {
	conf.Storage, rep, err = node.GetIgnitionStorage()
	conf.Ignition.Version = "3.1.0"
	return
}
