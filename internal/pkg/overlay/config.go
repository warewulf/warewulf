package overlay

import (
	"path"

	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
)

func OverlaySourceTopDir() string {
	return warewulfconf.Config("WWOVERLAYDIR")
}

func OverlaySourceDir(overlayName string) string {
	return path.Join(OverlaySourceTopDir(), overlayName)
}

func OverlayImage(nodeName string, overlayName string) string {
	return path.Join(warewulfconf.Config("WWPROVISIONDIR"), "overlays", nodeName, overlayName+".img")
}
