package container

import (
	"path"

	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
)

func SourceParentDir() string {
	conf := warewulfconf.New()
	return conf.WWCHROOTDIR()
}

func SourceDir(name string) string {
	return path.Join(SourceParentDir(), name)
}

func RootFsDir(name string) string {
	return path.Join(SourceDir(name), "rootfs")
}

func ImageParentDir() string {
	conf := warewulfconf.New()
	return path.Join(conf.WWPROVISIONDIR(), "container/")
}

func ImageFile(name string) string {
	return path.Join(ImageParentDir(), name+".img")
}
