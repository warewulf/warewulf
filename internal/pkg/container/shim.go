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
find the path of the shim binary in container
*/
func ShimFind(container string) string {
	container_path := RootFsDir(container)
	wwlog.Debug("Finding shim under path: %s", container_path)
	if container_path == "" {
		return ""
	}
	return ShimFindPath(container_path)
}

/*
find the path of the shim binary in container
*/
func ShimFindPath(shimpath string) string {
	for _, shimdir := range shimDirs() {
		wwlog.Debug("Checking shim directory: %s", shimdir)
		for _, shimname := range shimNames() {
			wwlog.Debug("Checking for shim name: %s", shimname)
			shimPaths, _ := filepath.Glob(path.Join(shimpath, shimdir, shimname))
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
