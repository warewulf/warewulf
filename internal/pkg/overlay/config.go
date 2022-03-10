package overlay

import (
	"path"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
)

func OverlaySourceTopDir() string {
	return buildconfig.WWOVERLAYDIR()
}

func OverlaySourceDir(overlayName string) string {
	return path.Join(OverlaySourceTopDir(), overlayName)
}

/*
Returns the overlay name of the image for a given node
*/
func OverlayImage(nodeName string, overlayName []string) string {
	return path.Join(buildconfig.WWPROVISIONDIR(), "overlays/", nodeName, strings.Join(overlayName, "-")+".img")
}
