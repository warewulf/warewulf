package configure

import (
	"fmt"

	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
)

/*
Creates '/etc/exports' from the host template, enables and start the
nfs server.
*/
func NFS() error {

	controller := warewulfconf.New()

	if controller.Nfs.Enabled {
		if controller.Warewulf.EnableHostOverlay {
			err := overlay.BuildHostOverlay()
			if err != nil {
				wwlog.Warn("host overlay could not be built: %s", err)
			}
		} else {
			wwlog.Info("host overlays are disabled, did not modify exports")
		}
		fmt.Printf("Enabling and restarting the NFS services\n")
		if controller.Nfs.SystemdName == "" {
			err := util.SystemdStart("nfs-server")
			if err != nil {
				return errors.Wrap(err, "failed to start nfs-server")
			}
		} else {
			err := util.SystemdStart(controller.Nfs.SystemdName)
			if err != nil {
				return errors.Wrap(err, "failed to start")
			}
		}
	}

	return nil
}
