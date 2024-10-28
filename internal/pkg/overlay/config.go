package overlay

import (
	"os"
	"path"
	"strings"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
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

// OverlayImage returns the full path to an overlay image based on the
// context and the overlays contained in it.
//
// If a context is provided, the image file name is based on that
// context name, in the form __{CONTEXT}__.
//
// If the context is empty ("") the image file name is a concatenated
// list of the contained overlays joined by "-".
//
// If the context is empty and no overlays are specified, the empty
// string is returned.
func OverlayImage(nodeName string, context string, overlayNames []string) string {
	var name string
	if context != "" {
		if len(overlayNames) > 0 {
			wwlog.Debug("context(%v) and overlays(%v) specified: prioritizing context(%v)",
				context, overlayNames, context)
		}
		name = "__" + strings.ToUpper(context) + "__.img"
	} else if len(overlayNames) > 0 {
		name = strings.Join(overlayNames, "-") + ".img"
	} else {
		wwlog.Warn("unable to generate overlay image path: no context or overlays specified")
		return ""
	}

	conf := warewulfconf.Get()
	return path.Join(conf.Paths.OverlayProvisiondir(), nodeName, name)
}
