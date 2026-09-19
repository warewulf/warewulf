package configure

import (
	"fmt"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
)

/*
Creates '/etc/hosts' by building the configured hostfile overlays.
*/
func Hostfile() error {
	var overlays warewulfconf.OverlayList
	if hostfile := warewulfconf.Get().Hostfile; hostfile != nil {
		overlays = hostfile.Overlays
	}
	if err := buildHostOverlays("hostfile", overlays); err != nil {
		return fmt.Errorf("could not build hostfile overlays: %w", err)
	}
	return nil
}
