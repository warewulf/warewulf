package container

import (
	"fmt"
	"path"

	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
)

var (
	initramfsSearchPaths = []string{
		// This is a printf format where the %s will be the kernel version
		"boot/initramfs-%s",
		"boot/initramfs-%s.img",
		"boot/initrd-%s",
		"boot/initrd-%s.img",
	}
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
func InitramfsBootPath(image, kver string) (string, error) {
	for _, searchPath := range initramfsSearchPaths {
		initramfs_path := path.Join(RootFsDir(image), fmt.Sprintf(searchPath, kver))
		wwlog.Debug("Looking for initramfs at: %s", initramfs_path)
		if util.IsFile(initramfs_path) {
			return initramfs_path, nil
		}
	}
	return "", fmt.Errorf("Failed to find a target kernel version initramfs")
}
