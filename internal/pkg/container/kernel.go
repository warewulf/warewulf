package container

import (
	"path"
	"path/filepath"

	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

var (
	kernelNames = []string{
		`vmlinux`,
		`vmlinuz`,
		`vmlinuz.gz`,
	}
	modulePath = "/lib/modules/"
)

func KernelFind(container string) string {
	wwlog.Printf(wwlog.DEBUG, "Finding kernel\n")
	container_path := RootFsDir(container)
	if container_path == "" {
		return ""
	}

	for _, kname := range kernelNames {
		wwlog.Printf(wwlog.DEBUG, "Checking for kernel name within module path: %s\n", kname)
		kernelPaths, err := filepath.Glob(path.Join(container_path, modulePath, "/*/", kname))
		if err != nil {
			return ""
		}
		for _, kernelPath := range kernelPaths {
			wwlog.Printf(wwlog.DEBUG, "Checking for kernel path: %s\n", kernelPath)
			if util.IsFile(kernelPath) {
				return kernelPath
			}
		}
	}

	return ""
}

func KernelVersion(container string) string {
	wwlog.Printf(wwlog.DEBUG, "Finding kernel version inside container: %s\n", container)
	kernel := KernelFind(container)
	if kernel == "" {
		return ""
	}

	return path.Base(path.Dir(kernel))
}
