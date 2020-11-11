package vnfs

import (
	"github.com/hpcng/warewulf/internal/pkg/config"
	"path"
	"strings"
)

type VnfsObject struct {
	SourcePath  string
	RootPath  	string
	ImagePath 	string
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

	return strings.ReplaceAll(uri[1], "/", "_")
}

func (self *VnfsObject) Source() string {
	if self.SourcePath == "" {
		return ""
	}

	return self.SourcePath
}


func (self *VnfsObject) Image() string {
	if self.SourcePath == "" {
		return ""
	}

	return config.LocalStateDir + "/provision/vnfs/" + self.NameClean() + ".img.gz"
}

func (self *VnfsObject) Root() string {
	if self.SourcePath == "" {
		return ""
	}

	return config.LocalStateDir + "/chroots/" + self.NameClean()
}
