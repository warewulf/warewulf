package configure

import (
	"fmt"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

/*
Creates '/etc/exports' from the host template, enables and start the
nfs server.
*/
func NFS() error {

	controller := warewulfconf.Get()

	if controller.NFS.Enabled() {
		if controller.Warewulf.EnableHostOverlay() {
			err := overlay.BuildHostOverlay()
			if err != nil {
				wwlog.Warn("host overlay could not be built: %s", err)
			}
		} else {
			wwlog.Info("host overlays are disabled, did not modify exports")
		}
		fmt.Printf("Enabling and restarting the NFS services\n")
		if controller.NFS.SystemdName == "" {
			err := util.SystemdStart("nfs-server")
			if err != nil {
				return fmt.Errorf("failed to start nfs-server: %w", err)
			}
		} else {
			err := util.SystemdStart(controller.NFS.SystemdName)
			if err != nil {
				return fmt.Errorf("failed to start: %w", err)
			}
		}
	}

	return nil
}
