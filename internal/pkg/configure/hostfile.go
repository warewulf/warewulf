package configure

import (
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

/*
Creates '/etc/hosts' by building the configured hostfile overlays.
*/
func Hostfile() (err error) {
	controller := warewulfconf.Get()

	if !controller.Warewulf.EnableHostOverlay() {
		wwlog.Info("host overlays are disabled, did not modify the hostfile")
		return nil
	}

	var overlays warewulfconf.OverlayList
	if controller.Hostfile != nil {
		overlays = controller.Hostfile.Overlays
	}
	if err := overlay.BuildHostOverlay(overlays...); err != nil {
		wwlog.Warn("host overlay could not be built: %s", err)
	}
	return nil
}
