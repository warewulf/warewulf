package overlay

import (
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/util"
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
	config := config.New()
	v := vnfs.New(vnfsname)

	vnfsdir := config.VnfsChroot(v.NameClean())

	if util.IsDir(vnfsdir) == false {
		wwlog.Printf(wwlog.WARN, "Template requesting file from non-imported VNFS: %s (%s)\n", vnfsname, filepath)
		return ""
	}
	wwlog.Printf(wwlog.DEBUG, "IncludeVnfs file from: %s/%s\n", vnfsdir, filepath)

	content, err := ioutil.ReadFile(path.Join(vnfsdir, filepath))

	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Template include: %s\n", err)
	}
	return strings.TrimSuffix(string(content), "\n")
}
