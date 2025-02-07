package kernel

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"

	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

var (
	kernelSearchPaths = []string{
		"/boot/Image-*", // this is the aarch64 for SUSE, vmlinux which is also present won't boot
		"/boot/vmlinuz-linux*",
		"/boot/vmlinuz-*",
		"/lib/modules/*/vmlinuz",
		"/lib/modules/*/vmlinuz.gz",
	}
)

type collection []*Kernel

func (k collection) Len() int {
	return len(k)
}

func (k collection) Less(i, j int) bool {
	iv := k[i].version()
	jv := k[j].version()
	return (iv == nil && jv != nil) ||
		(iv != nil && jv != nil && iv.LessThan(jv))
}

func (k collection) Swap(i, j int) {
	k[i], k[j] = k[j], k[i]
}

func (k collection) Sort() {
	sort.Sort(k)
}

func (k collection) Default() *Kernel {
	nk := append(collection{}, k...)
	sort.Sort(sort.Reverse(nk))
	for _, kernel := range nk {
		if !(kernel.IsDebug() || kernel.IsRescue()) {
			return kernel
		}
	}
	return nil
}

func (k collection) Version(version string) *Kernel {
	for _, kernel := range k {
		if kernel.IsDebug() || kernel.IsRescue() {
			continue
		} else if strings.HasPrefix(kernel.Version(), version) {
			return kernel
		}
	}
	return nil
}

type Kernel struct {
	Path      string
	ImageName string
}

func FromNode(node *node.Node) *Kernel {
	wwlog.Debug("FromNode(%v)", node)
	if node.ImageName == "" {
		return nil
	} else if node.Kernel != nil && node.Kernel.Version != "" {
		kernel := &Kernel{ImageName: node.ImageName, Path: filepath.Join("/", node.Kernel.Version)}
		if util.IsFile(kernel.FullPath()) {
			return kernel
		} else {
			return FindKernels(node.ImageName).Version(node.Kernel.Version)
		}
	} else {
		return FindKernels(node.ImageName).Default()
	}
}

func FindKernelsFromPattern(imageName string, pattern string) (kernels collection) {
	wwlog.Debug("FindKernelsFromPattern(%v, %v)", imageName, pattern)
	root := image.RootFsDir(imageName)
	fullPaths, err := filepath.Glob(filepath.Join(root, pattern))
	wwlog.Debug("%v: fullPaths: %v", filepath.Join(root, pattern), fullPaths)
	if err != nil {
		panic(err)
	}
	for _, fullPath := range fullPaths {
		path, err := filepath.Rel(root, fullPath)
		if err != nil {
			continue
		} else {
			kernels = append(kernels, &Kernel{ImageName: imageName, Path: filepath.Join("/", path)})
		}
	}
	return kernels
}

func FindKernels(imageName string) (kernels collection) {
	wwlog.Debug("FindKernels(%v)", imageName)
	for _, pattern := range kernelSearchPaths {
		kernels = append(kernels, FindKernelsFromPattern(imageName, pattern)...)
	}
	return kernels
}

func FindAllKernels() (kernels collection) {
	wwlog.Debug("FindAllKernels()")
	if sources, err := image.ListSources(); err == nil {
		for _, source := range sources {
			kernels = append(kernels, FindKernels(source)...)
		}
	} else {
		wwlog.Error("%s", err)
	}
	return kernels
}

func (kernel *Kernel) version() *version.Version {
	return util.ParseVersion(kernel.Path)
}

func (kernel *Kernel) Version() string {
	version := kernel.version()
	if version == nil {
		return ""
	} else {
		return version.String()
	}
}

func (kernel *Kernel) IsDebug() bool {
	return strings.Contains(kernel.Path, "+debug")
}

func (kernel *Kernel) IsRescue() bool {
	return strings.Contains(kernel.Path, "-rescue")
}

func (kernel *Kernel) FullPath() string {
	root := image.RootFsDir(kernel.ImageName)
	return filepath.Join(root, kernel.Path)
}
