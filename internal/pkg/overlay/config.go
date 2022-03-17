package overlay

import (
	"os"
	"path"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
)

func OverlaySourceTopDir() string {
	return buildconfig.WWOVERLAYDIR()
}

/*
Return the path for the base of the overlay, strips rootfs
prefix in the overlay dir if this it exists
*/
func OverlaySourceDir(overlayName string) string {
	/* Assume using old style overlay dir without rootfs */
	var overlaypath = path.Join(OverlaySourceTopDir(), overlayName)
	if _, err := os.Stat(path.Join(overlaypath, "rootfs")); err == nil {
		/* rootfs exists, use it. */
		overlaypath = path.Join(overlaypath, "rootfs")
	}
	return overlaypath
}

/*
Returns the overlay name of the image for a given node
*/
func OverlayImage(nodeName string, overlayName []string) string {
	return path.Join(buildconfig.WWPROVISIONDIR(), "overlays/", nodeName, strings.Join(overlayName, "-")+".img")
}
