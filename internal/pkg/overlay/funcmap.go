package overlay

import (
	"github.com/hpcng/warewulf/internal/pkg/container"
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

func templateContainerFileInclude(containername string, filepath string) string {
	wwlog.Printf(wwlog.DEBUG, "Including VNFS file into template: %s: %s\n", containername, filepath)

	if containername == "" {
		wwlog.Printf(wwlog.WARN, "VNFS not set for template import request: %s: %s\n", containername, filepath)
		return ""
	}

	if container.ValidSource(containername) == false {
		wwlog.Printf(wwlog.WARN, "Template required VNFS does not exist: %s\n", containername)
		return ""
	}

	containerDir := container.RootFsDir(containername)

	wwlog.Printf(wwlog.DEBUG, "Including file from container: %s:%s\n", containerDir, filepath)

	content, err := ioutil.ReadFile(path.Join(containerDir, filepath))

	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Template include: %s\n", err)
	}
	return strings.TrimSuffix(string(content), "\n")
}
