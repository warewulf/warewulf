package vnfs

import (
	"path"
	"strings"
)

type VnfsObject struct {
	SourcePath  string
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