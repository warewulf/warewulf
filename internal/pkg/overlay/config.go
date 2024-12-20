package overlay

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/containers/storage/drivers/copy"

	"github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// GetOverlay returns the filesystem path of an overlay identified by its name,
// along with a boolean indicating whether the returned overlayPath corresponds
// to a site-specific overlay.
func GetOverlay(name string) (overlay Overlay) {
	overlay = GetSiteOverlay(name)
	if overlay.Exists() {
		return overlay
	}
	overlay = GetDistributionOverlay(name)
	if overlay.Exists() {
		return overlay
	}
	return GetSiteOverlay(name)
}

// GetDistributionOverlay returns the filesystem path of a distribution overlay
// identified by the given name.
func GetDistributionOverlay(name string) (overlay Overlay) {
	return getOverlay(config.Get().Paths.DistributionOverlaydir(), name)
}

// GetSiteOverlay returns the filesystem path of a site-specific overlay
// identified by the given name.
func GetSiteOverlay(name string) (overlay Overlay) {
	return getOverlay(config.Get().Paths.SiteOverlaydir(), name)
}

// getOverlay constructs an overlay based on the given overlay directory and
// overlay name. The overlay does not necessarily exist.
func getOverlay(overlaydir, name string) (overlay Overlay) {
	return Overlay(path.Join(overlaydir, name))
}

// Create creates a new overlay directory for the given overlay
//
// Returns an error if the overlay already exists or if directory creation fails.
func (this Overlay) Create() error {
	if util.IsDir(this.Path()) {
		return fmt.Errorf("overlay already exists: %s", this)
	}
	return os.MkdirAll(this.Rootfs(), 0755)
}

// Creates a site overlay from an existing distribution overlay.
//
// If the distribution overlay doesn't exist, return an error.
func (this Overlay) CloneSiteOverlay() (siteOverlay Overlay, err error) {
	siteOverlay = GetSiteOverlay(this.Name())
	if !util.IsDir(this.Path()) {
		return siteOverlay, fmt.Errorf("source overlay does not exist: %s", this.Name())
	}
	if siteOverlay.Exists() {
		return siteOverlay, fmt.Errorf("site overlay already exists: %s", siteOverlay.Name())
	}
	err = copy.DirCopy(this.Path(), siteOverlay.Path(), copy.Content, true)
	return siteOverlay, err
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

	return path.Join(config.Get().Paths.OverlayProvisiondir(), nodeName, name)
}
