package configure

import (
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// buildHostOverlays builds the host overlays for a service, honouring
// the `warewulf:host overlay` setting. Returns an error only if a build
// fails.
func buildHostOverlays(service string, overlays warewulfconf.OverlayList) error {
	if !warewulfconf.Get().Warewulf.EnableHostOverlay() {
		wwlog.Info("host overlays are disabled, did not configure %s", service)
		return nil
	}
	if len(overlays) == 0 {
		wwlog.Warn("No host overlays configured for %s: set `%s:overlays` in warewulf.conf"+
			" or run `wwctl upgrade config`. %s configuration was not applied.",
			service, service, service)
		return nil
	}
	return overlay.BuildHostOverlay(overlays...)
}
