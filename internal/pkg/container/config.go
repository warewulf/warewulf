package container

import (
	"path"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
)

func SourceParentDir() string {
	conf := warewulfconf.Get()
	return conf.Paths.WWChrootdir
}

func SourceDir(name string) string {
	return path.Join(SourceParentDir(), name)
}

func RootFsDir(name string) string {
	return path.Join(SourceDir(name), "rootfs")
}

func ImageParentDir() string {
	conf := warewulfconf.Get()
	return path.Join(conf.Paths.WWProvisiondir, "container/")
}

func ImageFile(name string) string {
	return path.Join(ImageParentDir(), name+".img")
}

// InitramfsBootPath returns the dracut built initramfs path, as dracut built initramfs inside container
// the function returns host path of the built file
func InitramfsBootPath(image, kver string) string {
	return path.Join(RootFsDir(image), "boot", "initramfs-"+kver+".img")
}
