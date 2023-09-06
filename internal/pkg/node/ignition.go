package node

import (
	"fmt"
	"sort"

	types_3_4 "github.com/coreos/ignition/v2/config/v3_2/types"
	"github.com/coreos/vcontext/path"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

/*
Create a ignition struct class for ignition
*/
func (node *NodeInfo) GetStorage() (stor types_3_4.Storage, err error, rep string) {
	var fileSystems []types_3_4.Filesystem
	for fsdevice, fs := range node.FileSystems {
		var mountOptions []types_3_4.MountOption
		for _, opt := range fs.MountOptions.GetSlice() {
			mountOptions = append(mountOptions, types_3_4.MountOption(opt))
		}
		var fsOption []types_3_4.FilesystemOption
		for _, opt := range fs.Options.GetSlice() {
			fsOption = append(fsOption, types_3_4.FilesystemOption(opt))
		}
		wipe := fs.WipeFileSystem.GetB()
		myFs := types_3_4.Filesystem{
			Device:         fsdevice,
			Path:           fs.Path.GetPointer(),
			WipeFilesystem: &wipe,
		}
		if fs.Format.Get() != "" {
			myFs.Format = fs.Format.GetPointer()
		}
		if fs.Label.Get() != "" {
			myFs.Label = fs.Label.GetPointer()
		}
		if fs.MountOptions.Get() != "" {
			myFs.MountOptions = mountOptions
		}
		if fs.Options.Get() != "" {
			myFs.Options = fsOption
		}
		if fs.Options.Get() != "" {
			myFs.UUID = fs.Uuid.GetPointer()
		}
		wwlog.Debug("created file system struct: %v", myFs)
		fileSystems = append(fileSystems, myFs)
	}
	var disks []types_3_4.Disk
	for diskDev, disk := range node.Disks {
		var partitions []types_3_4.Partition
		for partlabel, part := range disk.Partitions {
			resize := part.Resize.GetB()
			shouldExist := part.ShouldExist.GetB()
			wipe := part.WipePartitionEntry.GetB()
			label := partlabel
			myPart := types_3_4.Partition{
				Label:              &label,
				Number:             part.Number.GetInt(),
				ShouldExist:        &shouldExist,
				WipePartitionEntry: &wipe,
			}
			if part.Guid.Get() != "" {
				myPart.GUID = part.Guid.GetPointer()
			}
			if part.Resize.Get() != "" {
				myPart.Resize = &resize
			}
			if part.SizeMiB.Get() != "" {
				myPart.SizeMiB = part.SizeMiB.GetIntPtr()
			}
			if part.StartMiB.Get() != "" {
				myPart.StartMiB = part.StartMiB.GetIntPtr()
			}
			if part.TypeGuid.Get() != "" {
				myPart.TypeGUID = part.TypeGuid.GetPointer()
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
		wipe := disk.WipeTable.GetB()
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
		err = fmt.Errorf(report.String())
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
func (node *NodeInfo) GetConfig() (conf SimpleIgnitionConfig, rep string, err error) {
	conf.Storage, err, rep = node.GetStorage()
	conf.Ignition.Version = "3.1.0"
	return
}
