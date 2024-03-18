package container

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/warewulf/warewulf/internal/pkg/kernel"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func KernelFind(container string) string {
	wwlog.Debug("Finding kernel")
	container_path := RootFsDir(container)
	if container_path == "" {
		return ""
	}

	for _, searchPath := range kernel.KernelSearchPaths {
		testPath := fmt.Sprintf(path.Join(container_path, searchPath), "*")
		wwlog.Debug("Checking for kernel name: %s", testPath)

		kernelPaths, err := filepath.Glob(testPath)
		if err != nil {
			return ""
		}

		if len(kernelPaths) == 0 {
			continue
		}

		sort.Slice(kernelPaths, func(i, j int) bool {
			return kernelPaths[i] > kernelPaths[j]
		})

		for _, kernelPath := range kernelPaths {
			wwlog.Debug("Checking for kernel path: %s", kernelPath)
			// Only succeeds if kernelPath exists and, if a
			// symlink, links to a path that also exists
			kernelPath, err = filepath.EvalSymlinks(kernelPath)
			if err == nil {
				wwlog.Debug("found kernel: %s", kernelPath)
				return kernelPath
			}
		}
	}

	return ""
}

func KernelVersion(container string) string {
	wwlog.Debug("Finding kernel version inside container: %s", container)
	kernel := KernelFind(container)
	if kernel == "" {
		return ""
	}

	ret := path.Base(path.Dir(kernel))
	if ret == "boot" {
		ret = path.Base(kernel)
	}

	return strings.TrimPrefix(ret, "vmlinuz-")
}
