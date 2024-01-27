package container

import (
	"os"
	"path"
	"path/filepath"

	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func shimDirs() []string {
	return []string{
		`/usr/share/efi/x86_64/`,
		`/usr/lib64/efi/`,
		`/boot/efi/EFI/*/`,
	}
}
func shimNames() []string {
	return []string{
		`shim.efi`,
		`shim-sles.efi`,
		`shimx64.efi`,
		`shim-susesigned.efi`,
	}
}

func grubDirs() []string {
	return []string{
		`/usr/lib64/efi/`,
		`/usr/share/grub2/*-efi/`,
		`/usr/share/efi/*/`,
		`/boot/efi/EFI/*/`,
	}
}
func grubNames() []string {
	return []string{
		`grub-tpm.efi`,
		`grub.efi`,
		`grubx64.efi`,
		`grubia32.efi`,
		`grubaa64.efi`,
		`grubarm.efi`,
	}
}

/*
find the path of the shim binary in container
*/
func ShimFind(container string) string {
	var container_path string
	if container != "" {
		container_path = RootFsDir(container)
	} else {
		container_path = "/"
	}
	wwlog.Debug("Finding grub under paths: %s", container_path)
	return BootLoaderFindPath(container_path, shimNames, shimDirs)
}

/*
find a grub.efi in the used container
*/
func GrubFind(container string) string {
	var container_path string
	if container != "" {
		container_path = RootFsDir(container)
	} else {
		container_path = "/"
	}
	wwlog.Debug("Finding grub under paths: %s", container_path)
	return BootLoaderFindPath(container_path, grubNames, grubDirs)
}

/*
find the path of the shim binary in container
*/
func BootLoaderFindPath(cpath string, names func() []string, paths func() []string) string {
	for _, bdir := range paths() {
		wwlog.Debug("Checking shim directory: %s", bdir)
		for _, bname := range names() {
			wwlog.Debug("Checking for bootloader name: %s", bname)
			shimPaths, _ := filepath.Glob(path.Join(cpath, bdir, bname))
			for _, shimPath := range shimPaths {
				wwlog.Debug("Checking for bootloader path: %s", shimPath)
				// Only succeeds if shimPath exists and, if a
				// symlink, links to a path that also exists
				if _, err := os.Stat(shimPath); err == nil {
					return shimPath
				}
			}
		}
	}
	return ""
}
