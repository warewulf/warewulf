package configure

import (
	"fmt"

	"github.com/pkg/errors"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

/*
Configures the dhcpd server, when show is set to false, else the
dhcp configuration is checked.
*/
func DHCP() (err error) {

	controller := warewulfconf.Get()

	if !controller.DHCP.Enabled {
		wwlog.Warn("This system is not configured as a Warewulf DHCP controller")
		return
	}

	if controller.DHCP.RangeStart == "" {
		return fmt.Errorf("configuration is not defined: `dhcpd range start`")
	}

	if controller.DHCP.RangeEnd == "" {
		return fmt.Errorf("configuration is not defined: `dhcpd range end`")
	}
	if controller.Warewulf.EnableHostOverlay {
		err = overlay.BuildHostOverlay()
		if err != nil {
			wwlog.Warn("host overlay could not be built: %s", err)
		}
	} else {
		wwlog.Info("host overlays are disabled, did not modify/create dhcpd configuration")
	}
	fmt.Printf("Enabling and restarting the DHCP services\n")
	err = util.SystemdStart(controller.DHCP.SystemdName)
	if err != nil {
		return errors.Wrap(err, "failed to start")
	}

	return
}
