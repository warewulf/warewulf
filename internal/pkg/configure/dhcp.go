package configure

import (
	"fmt"
	"os"

	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
)

/*
Configures the dhcpd server, when show is set to false, else the
dhcp configuration is checked.
*/
func Dhcp() error {

	controller, err := warewulfconf.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	if !controller.Dhcp.Enabled {
		wwlog.Printf(wwlog.INFO, "This system is not configured as a Warewulf DHCP controller\n")
		os.Exit(1)
	}

	if controller.Dhcp.RangeStart == "" {
		wwlog.Printf(wwlog.ERROR, "Configuration is not defined: `dhcpd range start`\n")
		os.Exit(1)
	}

	if controller.Dhcp.RangeEnd == "" {
		wwlog.Printf(wwlog.ERROR, "Configuration is not defined: `dhcpd range end`\n")
		os.Exit(1)
	}
	if controller.Warewulf.EnableHostOverlay {
		err = overlay.BuildHostOverlay()
		if err != nil {
			wwlog.Printf(wwlog.WARN, "host overlay could not be built: %s\n", err)
		}
	} else {
		wwlog.Printf(wwlog.INFO, "host overlays are disabled, did not modify/create dhcpd configuration")
	}
	fmt.Printf("Enabling and restarting the DHCP services\n")
	err = util.SystemdStart(controller.Dhcp.SystemdName)
	if err != nil {
		return errors.Wrap(err, "failed to start")
	}

	return nil
}
