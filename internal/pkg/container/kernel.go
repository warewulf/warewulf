package container

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"

	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

var (
	kernelSearchPaths = []string{
		// This is a printf format where the %s will be the kernel version
		`/boot/vmlinuz-%s`,
		`/boot/vmlinuz-%s.gz`,
		`/lib/modules/%s/vmlinuz`,
		`/lib/modules/%s/vmlinuz.gz`,
	}
)

func KernelFind(container string) string {
	container_path := RootFsDir(container)
	if container_path == "" {
		return ""
	}

	kernelVersion := KernelVersion(container)
	if kernelVersion == "" {
		return ""
	}

	for _, searchPath := range kernelSearchPaths {
		testPath := fmt.Sprintf(searchPath, kernelVersion)
		wwlog.Printf(wwlog.VERBOSE, "Looking for kernel at: '%s'\n", testPath)
		if util.IsFile(path.Join(container_path, testPath)) {
			return path.Join(container_path, testPath)
		}
	}

	return ""
}

func KernelVersion(container string) string {
	container_path := RootFsDir(container)
	if container_path == "" {
		return ""
	}

	module_lib_path := path.Join(container_path, "/lib/modules/*")
	wwlog.Printf(wwlog.DEBUG, "Searching for kernel modules at: %s\n", module_lib_path)
	kernelversions, err := filepath.Glob(module_lib_path)
	if err != nil {
		return ""
	}

	if len(kernelversions) > 1 {
		sort.Slice(kernelversions, func(i, j int) bool {
			return kernelversions[i] > kernelversions[j]
		})
		wwlog.Printf(wwlog.VERBOSE, "Multiple kernels found in container: %s\n", container)

		wwlog.Printf(wwlog.DEBUG, "Found lib path: '%s'\n", kernelversions[0])
		return path.Base(kernelversions[0])
	} else if len(kernelversions) == 1 {
		wwlog.Printf(wwlog.DEBUG, "Found lib path: '%s'\n", kernelversions[0])

		return path.Base(kernelversions[0])
	}

	return ""
}
