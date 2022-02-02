package container

import (
	"path"

	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
)

func SourceParentDir() string {
	return warewulfconf.Config("WWCHROOTDIR")
}

func SourceDir(name string) string {
	return path.Join(SourceParentDir(), name)
}

func RootFsDir(name string) string {
	return path.Join(SourceDir(name), "rootfs")
}

func ImageParentDir() string {
	return path.Join(warewulfconf.Config("WWPROVISIONDIR"), "container")
}

func ImageFile(name string) string {
	return path.Join(ImageParentDir(), name+".img.gz")
}
