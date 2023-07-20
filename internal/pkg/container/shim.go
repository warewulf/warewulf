package container

import (
	"os"
	"path"
	"path/filepath"

	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

func shimDirs() []string {
	return []string{
		`/usr/share/efi/x86_64/`,
		`/usr/lib64/efi`,
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

/*
find the path of the shim binary
*/
func ShimFind(container string) string {
	wwlog.Debug("Finding shim")
	container_path := RootFsDir(container)
	if container_path == "" {
		return ""
	}
	for _, shimdir := range shimDirs() {
		wwlog.Debug("Checking shim directory: %s", shimdir)
		for _, shimname := range shimNames() {
			wwlog.Debug("Checking for shim name: %s", shimname)
			shimPaths, _ := filepath.Glob(path.Join(container_path, shimdir, shimname))
			for _, shimPath := range shimPaths {
				wwlog.Debug("Checking for shim path: %s", shimPath)
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
