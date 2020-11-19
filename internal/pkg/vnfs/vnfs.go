package vnfs

import (
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"path"
	"strings"
)

type VnfsObject struct {
	SourcePath string
}

func New(s string) VnfsObject {
	var ret VnfsObject

	ret.SourcePath = s

	return ret
}

func (self *VnfsObject) Name() string {
	if self.SourcePath == "" {
		return ""
	}

	if strings.HasPrefix(self.SourcePath, "/") {
		return path.Base(self.SourcePath)
	}

	return self.SourcePath
}

func (self *VnfsObject) NameClean() string {
	if self.SourcePath == "" {
		return ""
	}

	if strings.HasPrefix(self.SourcePath, "/") {
		return path.Base(self.SourcePath)
	}
	uri := strings.Split(self.SourcePath, "://")

	return strings.ReplaceAll(uri[0]+":"+uri[1], "/", "_")
}

func (self *VnfsObject) Source() string {
	if self.SourcePath == "" {
		return ""
	}

	return self.SourcePath
}

func Build(uri string, force bool) error {
	v := New(uri)

	wwlog.Printf(wwlog.VERBOSE, "Building VNFS: %s\n", uri)
	if strings.HasPrefix(uri, "/") {
		if strings.HasSuffix(uri, "tar.gz") {
			//wwlog.Printf(wwlog.WARN, "Building VNFS from local tarball: %s\n", uri)
			wwlog.Printf(wwlog.WARN, "Building VNFS from local tarball is not supported yet: %s\n", uri)
		} else {
			BuildContainerdir(v, force)
		}
	} else {
		BuildDocker(v, force)
	}

	return nil
}
