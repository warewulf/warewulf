package overlay

import (
	"io/ioutil"
	"path"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

func templateFileInclude(path string) string {
	wwlog.Printf(wwlog.DEBUG, "Including file into template: %s\n", path)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		wwlog.Printf(wwlog.VERBOSE, "Could not include file into template: %s\n", err)
	}
	return strings.TrimSuffix(string(content), "\n")
}

func templateContainerFileInclude(containername string, filepath string) string {
	wwlog.Printf(wwlog.VERBOSE, "Including file from Container into template: %s:%s\n", containername, filepath)

	if containername == "" {
		wwlog.Printf(wwlog.WARN, "Container is not defined for node: %s\n", filepath)
		return ""
	}

	if !container.ValidSource(containername) {
		wwlog.Printf(wwlog.WARN, "Template requires file(s) from non-existant container: %s:%s\n", containername, filepath)
		return ""
	}

	containerDir := container.RootFsDir(containername)

	wwlog.Printf(wwlog.DEBUG, "Including file from container: %s:%s\n", containerDir, filepath)

	if !util.IsFile(path.Join(containerDir, filepath)) {
		wwlog.Printf(wwlog.WARN, "Requested file from container does not exist: %s:%s\n", containername, filepath)
		return ""
	}

	content, err := ioutil.ReadFile(path.Join(containerDir, filepath))

	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Template include failed: %s\n", err)
	}
	return strings.TrimSuffix(string(content), "\n")
}
