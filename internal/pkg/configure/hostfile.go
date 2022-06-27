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
	if !(util.IsFile(path.Join(overlay.OverlaySourceDir("host"), "/host/etc/hosts.ww"))) {
		wwlog.Error("'the overlay template '/etc/hosts.ww' does not exists in 'host' overlay\n")
		os.Exit(1)
	}
	var nodeInfo node.NodeInfo
	tstruct := overlay.InitStruct(nodeInfo)
	hostname, _ := os.Hostname()
	nodeInfo.Id.Set(hostname)
	buffer, backupFile, writeFile, err := overlay.RenderTemplateFile(
		path.Join(overlay.OverlaySourceDir("host"), "/host/etc/hosts.ww"),
		tstruct)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}
	info, err := os.Stat(path.Join(overlay.OverlaySourceDir("host"), "/host/etc/hosts.ww"))
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
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
