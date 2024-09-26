package build

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	controller := warewulfconf.Get()
	nodeDB, err := node.New()
	if err != nil {
		return fmt.Errorf("could not open node configuration: %s", err)
	}

	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		return fmt.Errorf("could not get node list: %s", err)
	}

	if len(args) > 0 {
		args = hostlist.Expand(args)
		nodes = node.FilterByName(nodes, args)

		if len(nodes) < len(args) {
			return errors.New("failed to find nodes")
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
			return errors.New("must specify overlay(s) to build")
		}

		if len(args) > 0 {
			if len(nodes) != 1 {
				return errors.New("must specify one node to build overlay")
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
			return fmt.Errorf("host overlay could not be built: %s", err)
		}
	}

	if BuildNodes || (!BuildHost && !BuildNodes) {
		if len(OverlayNames) > 0 {
			err = overlay.BuildSpecificOverlays(nodes, OverlayNames)
		} else {
			err = overlay.BuildAllOverlays(nodes)
		}

		if err != nil {
			return fmt.Errorf("some overlays failed to be generated: %s", err)
		}
	}
	return nil
}
