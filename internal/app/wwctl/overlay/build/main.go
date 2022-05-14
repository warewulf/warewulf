package build

import (
	"errors"
	"os"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/hpcng/warewulf/pkg/hostlist"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	controller, err := warewulfconf.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}
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
	if OverlayDir != "" {
		if OverlayName == "" {
			return errors.New("no overlay name given")
		}
		if len(args) > 0 {
			args = hostlist.Expand(args)
			for _, node := range nodes {
				if util.InSlice(node.RuntimeOverlay.GetSlice(), OverlayName) ||
					util.InSlice(node.SystemOverlay.GetSlice(), OverlayName) {
					return overlay.BuildOverlayIndir(node, strings.Split(OverlayName, ","), OverlayDir)
				} else {
					return errors.New("no node uses the given overlay")
				}
			}
		} else {
			var host node.NodeInfo
			var idEntry node.Entry
			hostname, _ := os.Hostname()
			wwlog.Printf(wwlog.INFO, "Building overlay for %s: host\n", hostname)
			idEntry.Set(hostname)
			host.Id = idEntry
			return overlay.BuildOverlayIndir(host, strings.Split(OverlayName, ","), OverlayDir)

		}

	}
	if BuildHost || (!BuildHost && !BuildNodes && len(args) == 0 && controller.Warewulf.EnableHostOverlay) {
		err := overlay.BuildHostOverlay()
		if err != nil {
			wwlog.Printf(wwlog.WARN, "host overlay could not be built: %s\n", err)
		}
	}
	if BuildNodes || (!BuildHost && !BuildNodes) {

		if len(args) > 0 {
			args = hostlist.Expand(args)
			if OverlayName != "" {
				err = overlay.BuildSpecificOverlays(node.FilterByName(nodes, args), OverlayName)
			} else {
				err = overlay.BuildAllOverlays(node.FilterByName(nodes, args))
			}
		} else {
			if OverlayName != "" {
				for _, n := range nodes {
					if util.InSlice(n.RuntimeOverlay.GetSlice(), OverlayName) ||
						util.InSlice(n.SystemOverlay.GetSlice(), OverlayName) {
						err = overlay.BuildSpecificOverlays([]node.NodeInfo{n}, OverlayName)
					}
				}
			} else {
				err = overlay.BuildAllOverlays(nodes)
			}
		}

		if err != nil {
			wwlog.Printf(wwlog.WARN, "Some system overlays failed to be generated: %s\n", err)

		}
	}
	return nil
}
