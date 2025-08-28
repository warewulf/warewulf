package overlay

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/containers/storage/drivers/copy"

	"github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// Get returns the filesystem path of an overlay identified by its name,
func Get(name string) (overlay Overlay, err error) {
	overlay = getSiteOverlay(name)
	if overlay.Exists() {
		return overlay, nil
	}
	overlay = getDistributionOverlay(name)
	if overlay.Exists() {
		return overlay, nil
	}
	return "", ErrDoesNotExist
}

// Create creates a new overlay directory for the given overlay
//
// Returns an error if the overlay already exists or if directory creation fails.
func Create(name string) (overlay Overlay, err error) {
	overlay = getSiteOverlay(name)
	if overlay.Exists() {
		return overlay, fmt.Errorf("overlay already exists: %s", name)
	}
	wwlog.Verbose("created site overlay under: %s", overlay.Path())
	return overlay, os.MkdirAll(path.Join(overlay.Path(), "rootfs"), 0o755)
}

// GetDistributionOverlay returns a distribution overlay identified by the given
// name.
func getDistributionOverlay(name string) Overlay {
	return Overlay(path.Join(config.Get().Paths.DistributionOverlaydir(), name))
}

// getSiteOverlay returns a site-specific overlay identified by the given name.
func getSiteOverlay(name string) (overlay Overlay) {
	return Overlay(path.Join(config.Get().Paths.SiteOverlaydir(), name))
}

// CloneToSite creates a site overlay from an existing distribution overlay.
//
// If the distribution overlay doesn't exist, return an error.
func (overlay Overlay) CloneToSite() (siteOverlay Overlay, err error) {
	wwlog.Verbose("Cloning to site overlay: %s", overlay.Name())
	siteOverlay = getSiteOverlay(overlay.Name())
	if siteOverlay.Exists() {
		return siteOverlay, nil
	}

	if !overlay.Exists() {
		return siteOverlay, fmt.Errorf("source overlay does not exist: %s", overlay.Name())
	}

	if !util.IsDir(filepath.Dir(siteOverlay.Path())) {
		if err := os.MkdirAll(filepath.Dir(siteOverlay.Path()), 0o755); err != nil {
			return siteOverlay, err
		}
	}
	err = copy.DirCopy(overlay.Path(), siteOverlay.Path(), copy.Content, true)
	return siteOverlay, err
}

// Image returns the full path to an overlay image based on the
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
func Image(nodeName string, context string, overlayNames []string) string {
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

func RemoveImage(nodeName string, context string, overlayNames []string) error {
	imagePath := Image(nodeName, context, overlayNames)
	if util.IsFile(imagePath) {
		if err := os.Remove(imagePath); err != nil {
			return fmt.Errorf("failed to remove overlay image: %w", err)
		}
	}
	compressedImagePath := imagePath + ".gz"
	if util.IsFile(compressedImagePath) {
		if err := os.Remove(compressedImagePath); err != nil {
			return fmt.Errorf("failed to remove compressed overlay image: %w", err)
		}
	}
	return nil
}
