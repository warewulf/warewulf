package overlay

import (
	"github.com/hpcng/warewulf/internal/pkg/vnfs"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"io/ioutil"
	"path"
	"strings"
)

func templateFileInclude(path string) string {
	wwlog.Printf(wwlog.DEBUG, "Including file into template: %s\n", path)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		wwlog.Printf(wwlog.WARN, "Could not include file into template: %s\n", err)
	}
	return strings.TrimSuffix(string(content), "\n")
}

func templateVnfsFileInclude(vnfsname string, filepath string) string {
	wwlog.Printf(wwlog.DEBUG, "Including VNFS file into template: %s: %s\n", vnfsname, filepath)

	if vnfsname == "" {
		wwlog.Printf(wwlog.WARN, "VNFS not set for template import request: %s: %s\n", vnfsname, filepath)
		return ""
	}

	if vnfs.ValidSource(vnfsname) == false {
		wwlog.Printf(wwlog.WARN, "Template required VNFS does not exist: %s\n", vnfsname)
		return ""
	}

	vnfsDir := vnfs.RootFsDir(vnfsname)

	wwlog.Printf(wwlog.DEBUG, "Including file from VNFS: %s:%s\n", vnfsDir, filepath)

	content, err := ioutil.ReadFile(path.Join(vnfsDir, filepath))

	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Template include: %s\n", err)
	}
	return strings.TrimSuffix(string(content), "\n")
}
