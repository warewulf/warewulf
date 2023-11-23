package configure

import (
	"fmt"
	"os"

	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
)

/*
Configures the dhcpd server, when show is set to false, else the
dhcp configuration is checked.
*/
func DHCP() (err error) {

	controller := warewulfconf.Get()

	if !controller.DHCP.Enabled {
		wwlog.Info("This system is not configured as a Warewulf DHCP controller")
		os.Exit(1)
	}

	if controller.DHCP.RangeStart == "" {
		wwlog.Error("Configuration is not defined: `dhcpd range start`")
		os.Exit(1)
	}

	if controller.DHCP.RangeEnd == "" {
		wwlog.Error("Configuration is not defined: `dhcpd range end`")
		os.Exit(1)
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
