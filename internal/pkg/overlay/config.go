package overlay

import (
	"os"
	"path"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/wwlog"
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
func OverlayImage(nodeName string, overlayName []string, img_context ...string) string {
	var name string
	var context string

	/* Check optional context argument. If missing, default to legacy. */
	if len(img_context) == 0 {
		context = "legacy"
	} else {
		context = img_context[0]
	}

	conf := warewulfconf.Get()

	switch context {
	case "legacy":
		name = strings.Join(overlayName, "-")+".img"
	default:
		wwlog.Warn("Context %s passed to OverlayImage(), using %s to build image name.", context, "__" + strings.ToUpper(context) + "__")
		name = "__" + strings.ToUpper(context) + "__.img"
	}

	return path.Join(conf.Paths.WWProvisiondir, "overlays/", nodeName, name)
}
