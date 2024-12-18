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

// GetOverlay returns the filesystem path of an overlay identified by its name,
// along with a boolean indicating whether the returned overlayPath corresponds
// to a site-specific overlay.
func GetOverlay(name string) (overlayPath string, isSite bool) {
	overlayPath = GetSiteOverlay(name)
	if _, err := os.Stat(overlayPath); err == nil {
		return overlayPath, true
	}
	overlayPath = GetDistributionOverlay(name)
	return overlayPath, false
}

// GetDistributionOverlay returns the filesystem path of a distribution overlay
// identified by the given name.
func GetDistributionOverlay(name string) (overlayPath string) {
	controller := warewulfconf.Get()
	return getOverlay(controller.Paths.DistributionOverlaydir(), name)
}

// GetSiteOverlay returns the filesystem path of a site-specific overlay
// identified by the given name.
func GetSiteOverlay(name string) (overlayPath string) {
	controller := warewulfconf.Get()
	return getOverlay(controller.Paths.SiteOverlaydir(), name)
}

// getOverlay constructs the filesystem path of an overlay based on the given
// overlay directory and overlay name.
//
// The returned path will include a "rootfs" directory if it exists.
func getOverlay(overlaydir, name string) (overlayPath string) {
	/* Assume using old style overlay dir without rootfs */
	overlayPath = path.Join(overlaydir, name)
	if _, err := os.Stat(path.Join(overlayPath, "rootfs")); err == nil {
		/* rootfs exists, use it. */
		overlayPath = path.Join(overlayPath, "rootfs")
	}
	return overlayPath
}

// CreateSiteOverlay creates a new site overlay directory with the specified name.
//
// This function constructs the path for the new overlay based on the site's overlay directory
// configuration and checks if the directory already exists. If the overlay already exists,
// it returns an error. Otherwise, it creates the necessary directory structure, including a
// rootfs directory.
//
// Parameters:
//   - overlayName: The name of the site overlay to be created.
//
// Returns:
//   - overlayPath: The full path to the created site overlay directory.
//   - err: An error if the overlay already exists or if directory creation fails.
func CreateSiteOverlay(overlayName string) (overlayPath string, err error) {
	controller := warewulfconf.Get()
	overlayPath = path.Join(controller.Paths.SiteOverlaydir(), overlayName)
	if util.IsDir(overlayPath) {
		return overlayPath, fmt.Errorf("overlay already exists: %s", overlayName)
	}
	overlayPath = path.Join(overlayPath, "rootfs")
	err = os.MkdirAll(overlayPath, 0755)
	return overlayPath, err
}

// Creates a site overlay from an existing distribution overlay.
//
// If the distribution overlay doesn't exist, return an OverlayDoesNotExist error.
func CloneSiteOverlay(name string) (err error) {
	controller := warewulfconf.Get()
	distroPath := path.Join(controller.Paths.DistributionOverlaydir(), name)
	sitePath := path.Join(controller.Paths.SiteOverlaydir(), name)
	if !util.IsDir(distroPath) {
		return &OverlayDoesNotExist{Name: name}
	}
	err = copy.DirCopy(distroPath, sitePath, copy.Content, true)
	return err
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
