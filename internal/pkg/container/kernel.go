package container

import (
	"path"
	"path/filepath"
	"sort"

	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

var (
	kernelSearchPaths = []string{
		// This is a printf format where the %s will be the kernel version
		"/boot/vmlinuz-*",
		"/lib/modules/*/vmlinuz*",
	}
)

func KernelFind(container string) string {
	container_path := RootFsDir(container)
	if container_path == "" {
		return ""
	}

	for _, searchPath := range kernelSearchPaths {

		check_path := path.Join(container_path, searchPath)

		wwlog.Printf(wwlog.DEBUG, "Searching for kernel(s) at: %s\n", check_path)
		kernels, err := filepath.Glob(check_path)
		if err != nil {
			return ""
		}

		if len(kernels) > 1 {
			sort.Slice(kernels, func(i, j int) bool {
				return kernels[i] > kernels[j]
			})
			wwlog.Printf(wwlog.VERBOSE, "Multiple kernels found in container: %s\n", container)
			return kernels[0]
		} else if len(kernels) == 1 {
			return kernels[0]
		}

	}
	return "NOT_FOUND"
}
