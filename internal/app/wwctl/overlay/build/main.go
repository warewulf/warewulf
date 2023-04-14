package build

import (
	"errors"
	"os"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/hpcng/warewulf/pkg/hostlist"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	controller := warewulfconf.New()
	nodeDB, err := node.New()
	if err != nil {
		wwlog.Error("Could not open node configuration: %s", err)
		os.Exit(1)
	}

	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		wwlog.Error("Could not get node list: %s", err)
		os.Exit(1)
	}

	if len(args) > 0 {
		args = hostlist.Expand(args)
		nodes = node.FilterByName(nodes, args)

		if len(nodes) < len(args) {
			return errors.New("Failed to find nodes")
		}
	}

	// NOTE: this is to keep backward compatible
	// passing -O a,b,c versus -O a -O b -O c, but will also accept -O a,b -O c
	overlayNames := []string{}
	for _, name := range OverlayNames {
		names := strings.Split(name, ",")
		overlayNames = append(overlayNames, names...)
	}
	OverlayNames = overlayNames

	if OverlayDir != "" {
		if len(OverlayNames) == 0 {
			// TODO: should this behave the same as OverlayDir == "", and build default
			// set to overlays?
			return errors.New("Must specify overlay(s) to build")
		}

		if len(args) > 0 {
			if len(nodes) != 1 {
				return errors.New("Must specify one node to build overlay")
			}

			for _, node := range nodes {
				return overlay.BuildOverlayIndir(node, OverlayNames, OverlayDir)
			}
		} else {
			// TODO this seems different than what is set in BuildHostOverlay
			var host node.NodeInfo
			var idEntry node.Entry
			hostname, _ := os.Hostname()
			wwlog.Info("Building overlay for host: %s", hostname)
			idEntry.Set(hostname)
			host.Id = idEntry
			return overlay.BuildOverlayIndir(host, OverlayNames, OverlayDir)

		}

	}

	if BuildHost && controller.Warewulf.EnableHostOverlay {
		err := overlay.BuildHostOverlay()
		if err != nil {
			wwlog.Warn("host overlay could not be built: %s", err)
		}
	}

	if BuildNodes || (!BuildHost && !BuildNodes) {
		if len(OverlayNames) > 0 {
			err = overlay.BuildSpecificOverlays(nodes, OverlayNames)
		} else {
			err = overlay.BuildAllOverlays(nodes)
		}

		if err != nil {
			wwlog.Warn("Some overlays failed to be generated: %s", err)
		}
	}
	return nil
}
