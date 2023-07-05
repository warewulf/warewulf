package container

import (
	"path"
	"path/filepath"
	"sort"

	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

func shimDirs() []string {
	return []string{
		`/usr/share/efi/x86_64/`,
		`/usr/lib64/`,
		``}
}
func shimNames() []string {
	return []string{
		`shim.efi`,
		`shim-sles.efi`,
	}
}

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
			shimPaths, err := filepath.Glob(path.Join(container_path, shimdir, shimname))
			if err != nil {
				return ""
			}
			if len(shimPaths) == 0 {
				continue
			}
			sort.Slice(shimPaths, func(i, j int) bool {
				return shimPaths[i] > shimPaths[j]
			})
			for _, shimPath := range shimPaths {
				wwlog.Debug("Checking for shim path: %s", shimPath)
				// Only succeeds if shimPath exists and, if a
				// symlink, links to a path that also exists
				shimPath, err = filepath.EvalSymlinks(shimPath)
				if err == nil {
					wwlog.Debug("found shim: %s", shimPath)
					return shimPath
				}
			}
		}
	}
	return ""
}
