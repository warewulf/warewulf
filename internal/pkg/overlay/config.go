package overlay

import (
	"path"

	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
)

func OverlaySourceTopDir() string {
	return buildconfig.WWOVERLAYDIR
}

func OverlaySourceDir(overlayName string) string {
	return path.Join(OverlaySourceTopDir(), overlayName)
}

func OverlayImage(nodeName string, overlayName string) string {
	return path.Join(buildconfig.WWPROVISIONDIR, "overlays/", nodeName, overlayName+".img")
}
