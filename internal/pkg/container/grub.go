package container

import (
	"os"
	"path"
	"path/filepath"

	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

func grubDirs() []string {
	return []string{
		`/usr/share/grub2/x86_64-efi`,
		`/usr/share/efi/x86_64/`,
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
find a grub.efi in the used container
*/
func GrubFind(container string) string {
	container_path := RootFsDir(container)
	wwlog.Debug("Finding grub under paths: %s", container_path)
	if container_path == "" {
		return ""
	}
	return GrubFindPath(container_path)
}

func GrubFindPath(grub_path string) string {
	for _, grubdir := range grubDirs() {
		wwlog.Debug("Checking grub directory: %s", grubdir)
		for _, grubname := range grubNames() {
			wwlog.Debug("Checking for grub name: %s", grubname)
			grubPaths, _ := filepath.Glob(path.Join(grub_path, grubdir, grubname))
			for _, grubpath := range grubPaths {
				wwlog.Debug("Checking for grub path: %s", grubpath)
				// Only succeeds if grubpath exists and, if a
				// symlink, links to a path that also exists
				if _, err := os.Stat(grubpath); err == nil {
					return grubpath
				}
			}
		}
	}
	return ""
}
