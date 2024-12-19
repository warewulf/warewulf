package build

import (
	"errors"
	"fmt"
	"strings"
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

	db, err := nodeDB.FindAllNodes()
	if err != nil {
		return fmt.Errorf("could not get node list: %s", err)
	}

	if len(args) > 0 {
		args = hostlist.Expand(args)
		db = node.FilterNodeListByName(db, args)

		if len(db) < len(args) {
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
			if len(db) != 1 {
				return errors.New("must specify one node to build overlay")
			}

			for _, node := range db {
				return overlay.BuildOverlayIndir(node, OverlayNames, OverlayDir)
			}
		} else {
			return errors.New("must specify a node to build overlay")
		}

	}

	oldMask := syscall.Umask(007)
	defer syscall.Umask(oldMask)

	if len(OverlayNames) > 0 {
		err = overlay.BuildSpecificOverlays(db, OverlayNames, Workers)
	} else {
		err = overlay.BuildAllOverlays(db, Workers)
	}

	if err != nil {
		return fmt.Errorf("some overlays failed to be generated: %s", err)
	}
	return nil
}
