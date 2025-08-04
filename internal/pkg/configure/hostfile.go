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
	overlay_, err := overlay.GetOverlay("host")
	if err != nil {
		return err
	}
	hostTemplate := path.Join(overlay_.Rootfs(), "/etc/hosts.ww")
	if !(util.IsFile(hostTemplate)) {
		return fmt.Errorf("'the overlay template '/etc/hosts.ww' does not exists in 'host' overlay")
	}

	var allNodes []node.Node
	if nodeDB, err := node.New(); err != nil {
		return err
	} else {
		allNodes, err = nodeDB.FindAllNodes()
		if err != nil {
			return err
		}
	}

	hostname, _ := os.Hostname()
	tstruct, err := overlay.InitStruct(overlay_.Name(), node.NewNode(hostname), allNodes)
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
