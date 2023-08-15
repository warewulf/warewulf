package configure

import (
	"os"
	"path"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
)

/*
Creates '/etc/hosts' from the host template.
*/
func Hostfile() error {
	hostTemplate := path.Join(overlay.OverlaySourceDir("host"), "/etc/hosts.ww")
	if !(util.IsFile(hostTemplate)) {
		wwlog.Error("'the overlay template '/etc/hosts.ww' does not exists in 'host' overlay")
		os.Exit(1)
	}

	nodeInfo := node.NewInfo()
	hostname, _ := os.Hostname()
	nodeInfo.Id.Set(hostname)
	tstruct := overlay.InitStruct(&nodeInfo)
	buffer, backupFile, writeFile, err := overlay.RenderTemplateFile(
		hostTemplate,
		tstruct)
	if err != nil {
		wwlog.Error("%s", err)
		os.Exit(1)
	}
	info, err := os.Stat(hostTemplate)
	if err != nil {
		wwlog.Error("%s", err)
		os.Exit(1)
	}

	if writeFile {
		err = overlay.CarefulWriteBuffer("/etc/hosts", buffer, backupFile, info.Mode())
		if err != nil {
			return errors.Wrap(err, "could not write file from template")
		}
	}
	return nil
}
