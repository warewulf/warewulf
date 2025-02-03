package image

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"

	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

var (
	initramfsSearchPaths = []string{
		"/boot/initramfs-*",
		"/boot/initrd-*",
	}

	versionPattern *regexp.Regexp
)

func init() {
	versionPattern = regexp.MustCompile(`\d+\.\d+\.\d+(-[\d\.]+|)`)
}

type Initramfs struct {
	Path      string
	imageName string
}

func (this *Initramfs) version() *version.Version {
	matches := versionPattern.FindAllString(this.Path, -1)
	for i := len(matches) - 1; i >= 0; i-- {
		if version_, err := version.NewVersion(strings.TrimSuffix(matches[i], ".")); err == nil {
			return version_
		}
	}
	return nil
}

func (this *Initramfs) Version() string {
	version := this.version()
	if version == nil {
		return ""
	} else {
		return version.String()
	}
}

func (this *Initramfs) FullPath() string {
	root := RootFsDir(this.imageName)
	return filepath.Join(root, this.Path)
}

func FindInitramfsFromPattern(imageName string, version string, pattern string) (initramfs *Initramfs) {
	wwlog.Debug("FindInitramfsFromPattern(%v, %v, %v)", imageName, version, pattern)
	root := RootFsDir(imageName)
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
			initramfs := &Initramfs{Path: filepath.Join("/", path), imageName: imageName}
			if strings.HasPrefix(initramfs.Version(), version) {
				return initramfs
			}
		}
	}
	return nil
}

// FindInitramfs returns the Initramfs for a given image and (kernel) version
func FindInitramfs(imageName string, version string) *Initramfs {
	for _, pattern := range initramfsSearchPaths {
		initramfs := FindInitramfsFromPattern(imageName, version, pattern)
		if initramfs != nil {
			return initramfs
		}
	}
	return nil
}
