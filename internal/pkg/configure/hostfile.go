package configure

import (
	"fmt"
	"os"
	"path"

	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/util"
)

/*
Creates '/etc/hosts' from the host template.
*/
func Hostfile() (err error) {
	overlaySourceDir, _ := overlay.GetOverlay("host")
	hostTemplate := path.Join(overlaySourceDir, "/etc/hosts.ww")
	if !(util.IsFile(hostTemplate)) {
		return fmt.Errorf("'the overlay template '/etc/hosts.ww' does not exists in 'host' overlay")
	}

	hostname, _ := os.Hostname()
	tstruct, err := overlay.InitStruct(node.NewNode(hostname))
	if err != nil {
		return err
	}
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
			return fmt.Errorf("could not write file from template: %w", err)
		}
	}
	return
}
