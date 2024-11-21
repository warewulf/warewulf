package overlay

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/containers/storage/drivers/copy"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

/*
Return the path for the base of the overlay, adds rootfs
prefix in the overlay dir if this it exists
*/
func OverlaySourceDir(overlayName string) (overlaypath string, isSite bool) {
	controller := warewulfconf.Get()
	/* Assume using old style overlay dir without rootfs */
	overlaypath = path.Join(controller.Paths.Sysconfdir, "overlays", overlayName)
	if _, err := os.Stat(path.Join(overlaypath, "rootfs")); err == nil {
		/* rootfs exists, use it. */
		overlaypath = path.Join(overlaypath, "rootfs")
	}
	if _, err := os.Stat(overlaypath); err == nil {
		return overlaypath, true
	}
	overlaypath = path.Join(controller.Paths.WWOverlaydir, overlayName)
	if _, err := os.Stat(path.Join(overlaypath, "rootfs")); err == nil {
		/* rootfs exists, use it. */
		overlaypath = path.Join(overlaypath, "rootfs")
	}
	wwlog.Debug("found overlay %s in path: %s", overlayName, overlaypath)
	return overlaypath, false
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

type OverlayDoesNotExist struct {
	Name string
}

func (e *OverlayDoesNotExist) Error() string {
	return fmt.Sprintf("overlay %s does not exist", e.Name)
}

// Creates a site overlay from an existing overlay and give back
// OverlayDoesNotExist error if distribution overlay doesn't exsist
func CreateSiteOverlay(name string) (err error) {
	controller := warewulfconf.Get()
	distroPath := path.Join(controller.Paths.WWOverlaydir, name)
	sitePath := path.Join(controller.Paths.Sysconfdir, "overlays", name)
	if !util.IsDir(distroPath) {
		return &OverlayDoesNotExist{Name: name}
	}
	err = copy.DirCopy(distroPath, sitePath, copy.Content, true)
	return err
}
