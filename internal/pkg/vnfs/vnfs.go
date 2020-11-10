package vnfs

import (
	"github.com/hpcng/warewulf/internal/pkg/config"
	"path"
	"strings"
)

type VnfsObject struct {
	Source    string
	RootPath  string
	ImagePath string
}

func New(s string) VnfsObject {
	var ret VnfsObject

	ret.Source = s

	return ret
}

func (self *VnfsObject) Name() string {
	if self.Source == "" {
		return ""
	}

	if strings.HasPrefix(self.Source, "/") {
		return path.Base(self.Source)
	}

	uri := strings.Split(self.Source, "://")

	return strings.ReplaceAll(uri[0]+":"+uri[1], "/", "_")
}

func (self *VnfsObject) Image() string {
	if self.Source == "" {
		return ""
	}

	return config.LocalStateDir + "/provision/vnfs/" + self.Name() + ".img.gz"
}

func (self *VnfsObject) Root() string {
	if self.Source == "" {
		return ""
	}

	return config.LocalStateDir + "/chroots/" + self.Name()
}
