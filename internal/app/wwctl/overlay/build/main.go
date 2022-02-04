package build

import (
	"os"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/hpcng/warewulf/pkg/hostlist"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not get node list: %s\n", err)
		os.Exit(1)
	}

	if len(args) > 0 {
		args = hostlist.Expand(args)
		err = overlay.BuildAllOverlays(node.FilterByName(nodes, args))
	} else {
		err = overlay.BuildAllOverlays(nodes)
	}

	if err != nil {
		wwlog.Printf(wwlog.WARN, "Some system overlays failed to be generated: %s\n", err)
	}

	return nil
}
