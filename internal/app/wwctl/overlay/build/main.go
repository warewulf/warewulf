package build

import (
	"errors"
	"fmt"
	"runtime"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	nodeDB, err := node.New()
	if err != nil {
		return fmt.Errorf("could not open node configuration: %s", err)
	}

	allNodes, err := nodeDB.FindAllNodes()
	if err != nil {
		return fmt.Errorf("could not get node list: %s", err)
	}

	var filteredNodes []node.Node
	if len(args) > 0 {
		args = hostlist.Expand(args)
		filteredNodes = node.FilterNodeListByName(allNodes, args)

		if len(filteredNodes) < len(args) {
			return errors.New("failed to find nodes")
		}
	} else {
		filteredNodes = allNodes
	}

	oldMask := syscall.Umask(000)
	defer syscall.Umask(oldMask)

	workers := Workers
	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	if err = overlay.BuildAllOverlays(filteredNodes, allNodes, workers); err != nil {
		return fmt.Errorf("some overlays failed to be generated: %s", err)
	}
	return nil
}
