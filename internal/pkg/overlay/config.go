package overlay

import (
	"os"
	"path"
	"strings"

	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
)

func OverlaySourceTopDir() string {
	conf := warewulfconf.Get()
	return conf.Paths.WWOverlaydir
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
	conf := warewulfconf.Get()
	var name string
	if len(overlayName) == 1 {
		name = overlayName[0]
	} else {
		var sb strings.Builder
		for _, str := range overlayName {
			if len(str) > 0 {
				sb.WriteByte(str[0])
			}
		}
		name = "__wwmerged_" + sb.String()
	}
	return path.Join(conf.Paths.WWProvisiondir, "overlays/", nodeName, name + ".img")
}
