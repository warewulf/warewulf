package container

import (
	"path"
	"path/filepath"
	"sort"

	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

var (
	kernelNames = []string{
		`vmlinux`,
		`vmlinuz`,
		`vmlinux-*`,
		`vmlinuz-*`,
		`vmlinuz.gz` }

	kernelDirs = []string{
		`/lib/modules/*/`,
		`/boot/` }
)

func KernelFind(container string) string {
	wwlog.Debug("Finding kernel\n")
	container_path := RootFsDir(container)
	if container_path == "" {
		return ""
	}

	for _, kdir := range kernelDirs {
		wwlog.Debug("Checking kernel directory: %s\n", kdir)
		for _, kname := range kernelNames {
			wwlog.Debug("Checking for kernel name: %s\n", kname)
			kernelPaths, err := filepath.Glob(path.Join(container_path, kdir, kname))
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
				wwlog.Debug("Checking for kernel path: %s\n", kernelPath)
				if util.IsFile(kernelPath) {
					return kernelPath
				}
			}
		}
	}

	return ""
}

func KernelVersion(container string) string {
	wwlog.Debug("Finding kernel version inside container: %s\n", container)
	kernel := KernelFind(container)
	if kernel == "" {
		return ""
	}

	return path.Base(path.Dir(kernel))
}
