package container

import (
	"path"

	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
)

func SourceParentDir() string {
	return buildconfig.WWCHROOTDIR()
}

func SourceDir(name string) string {
	return path.Join(SourceParentDir(), name)
}

func RootFsDir(name string) string {
	return path.Join(SourceDir(name), "rootfs")
}

func ImageParentDir() string {
	return path.Join(buildconfig.WWPROVISIONDIR(), "container/")
}

func ImageFile(name string) string {
	return path.Join(ImageParentDir(), name+".img.gz")
}
