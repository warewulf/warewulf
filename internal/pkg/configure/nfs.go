package configure

import (
	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
)

/*
Creates '/etc/exports' from the host template, enables and start the
nfs server.
*/
func NFS() error {

	controller := warewulfconf.Get()

	if controller.NFS.Enabled {
		if controller.Warewulf.EnableHostOverlay {
			err := overlay.BuildHostOverlay()
			if err != nil {
				wwlog.Warn("host overlay could not be built: %s", err)
			}
		} else {
			wwlog.Info("host overlays are disabled, did not modify exports")
		}
		wwlog.Info("Enabling and restarting the NFS services")
		if controller.NFS.SystemdName == "" {
			err := util.SystemdStart("nfs-server")
			if err != nil {
				return errors.Wrap(err, "failed to start nfs-server")
			}
		} else {
			err := util.SystemdStart(controller.NFS.SystemdName)
			if err != nil {
				return errors.Wrap(err, "failed to start")
			}
		}
	}

	return nil
}
