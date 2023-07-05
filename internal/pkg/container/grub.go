package container

import (
	"path"
	"path/filepath"
	"sort"

	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

func grubDirs() []string {
	return []string{
		`/usr/share/grub2/x86_64-efi`,
		`/usr/share/efi/x86_64/`,
	}
}
func grubNames() []string {
	return []string{
		`grub-tpm.efi`,
		`grub.efi`,
	}
}

/*
Tries to find a grub.efi in the used container
*/
func GrubFind(container string) string {
	wwlog.Debug("Finding grub")
	container_path := RootFsDir(container)
	if container_path == "" {
		return ""
	}
	for _, grubdir := range grubDirs() {
		wwlog.Debug("Checking grub directory: %s", grubdir)
		for _, shimname := range shimNames() {
			wwlog.Debug("Checking for shim name: %s", shimname)
			grubPaths, err := filepath.Glob(path.Join(container_path, grubdir, shimname))
			if err != nil {
				return ""
			}
			if len(grubPaths) == 0 {
				continue
			}
			sort.Slice(grubPaths, func(i, j int) bool {
				return grubPaths[i] > grubPaths[j]
			})
			for _, grubPath := range grubPaths {
				wwlog.Debug("Checking for grub path: %s", grubPath)
				// Only succeeds if shimPath exists and, if a
				// symlink, links to a path that also exists
				grubPath, err = filepath.EvalSymlinks(grubPath)
				if err == nil {
					wwlog.Debug("found grub: %s", grubPath)
					return grubPath
				}
			}
		}
	}
	return ""
}
