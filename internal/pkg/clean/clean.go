package clean

import (
	"os"
	"path"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

/*
Cleans up the OCI cache and remains of deleted nodes
*/
func CleanOciBlobCacheDir() error {
	warewulfconf := warewulfconf.Get()
	wwlog.Verbose("removing oci cache dir: %s", warewulfconf.Paths.OciBlobCachedir())
	return os.RemoveAll(warewulfconf.Paths.OciBlobCachedir())
}

func CleanOverlays() error {
	warewulfconf := warewulfconf.Get()
	nodeDB, err := node.New()
	if err != nil {
		return err
	}
	nodes := nodeDB.ListAllNodes()
	dirList, err := os.ReadDir(warewulfconf.Paths.OverlayProvisiondir())
	if err != nil {
		return err
	}
	for _, item := range dirList {
		if !item.IsDir() {
			continue
		}
		if !util.InSlice(nodes, item.Name()) {
			wwlog.Verbose("removing overlays of delete node: %s", item.Name())
			err = os.RemoveAll(path.Join(warewulfconf.Paths.OverlayProvisiondir(), item.Name()))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
