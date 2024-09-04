package configure

import (
	"fmt"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/util"
)

/*
Creates '/etc/hosts' from the host template.
*/
func Hostfile() (err error) {
	hostTemplate := path.Join(overlay.OverlaySourceDir("host"), "/etc/hosts.ww")
	if !(util.IsFile(hostTemplate)) {
		return fmt.Errorf("'the overlay template '/etc/hosts.ww' does not exists in 'host' overlay")
	}

	nodeInfo := node.NewInfo()
	hostname, _ := os.Hostname()
	nodeInfo.Id.Set(hostname)
	tstruct := overlay.InitStruct(&nodeInfo)
	buffer, backupFile, writeFile, err := overlay.RenderTemplateFile(
		hostTemplate,
		tstruct)
	if err != nil {
		return
	}
	info, err := os.Stat(hostTemplate)
	if err != nil {
		return
	}

	if writeFile {
		err = overlay.CarefulWriteBuffer("/etc/hosts", buffer, backupFile, info.Mode())
		if err != nil {
			return errors.Wrap(err, "could not write file from template")
		}
	}
	return
}
