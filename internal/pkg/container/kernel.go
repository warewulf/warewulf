package container

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

var (
	kernelNames = []string{
		`vmlinux`,
		`vmlinuz`,
		`vmlinux-*`,
		`vmlinuz-*`,
		`vmlinuz.gz`}

	kernelDirs = []string{
		`/lib/modules/*/`,
		`/boot/`}
)

func KernelFind(container string) string {
	wwlog.Debug("Finding kernel")
	container_path := RootFsDir(container)
	if container_path == "" {
		return ""
	}

	for _, kdir := range kernelDirs {
		wwlog.Debug("Checking kernel directory: %s", kdir)
		for _, kname := range kernelNames {
			wwlog.Debug("Checking for kernel name: %s", kname)
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
				wwlog.Debug("Checking for kernel path: %s", kernelPath)
				if util.IsFile(kernelPath) {
					// IsFile does not check if file is a softlink
					stat, _ := os.Lstat(kernelPath)
					if stat.Mode()&os.ModeSymlink == os.ModeSymlink {
						wwlog.Verbose("%s is a softlink", kernelPath)
						kernelPath, err = filepath.EvalSymlinks(kernelPath)
						if err != nil {
							wwlog.Warn("could evaluate symlink %s: %s", kernelPath, err)
						}
						wwlog.Debug("softlink is %s", kernelPath)
					}
					return kernelPath
				}
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
