package clean

import (
	"os"
	"path/filepath"
	"strings"

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

	// Clean the base directory path FIRST
	baseDir := filepath.Clean(warewulfconf.Paths.OverlayProvisiondir())

	dirList, err := os.ReadDir(baseDir)
	if err != nil {
		return err
	}
	for _, item := range dirList {
		if !item.IsDir() {
			continue
		}

		if !util.InSlice(nodes, item.Name()) {

			// Construct and validate the path (filepath.Join already calls Clean)
			cleanTarget := filepath.Join(baseDir, item.Name())

			// Verify the path is within baseDir
			rel, err := filepath.Rel(baseDir, cleanTarget)
			if err != nil {
				wwlog.Warn("failed to compute relative path for %s: %v", item.Name(), err)
				continue
			}

			if strings.HasPrefix(rel, "..") {
				wwlog.Warn("skipping directory with path traversal attempt: %s", item.Name())
				continue
			}

			wwlog.Verbose("removing overlays of deleted node: %s", item.Name())
			err = os.RemoveAll(cleanTarget)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
