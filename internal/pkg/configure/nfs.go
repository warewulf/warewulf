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
Creates '/etc/exports' from the host template, enables and start the
nfs server.
*/
func NFS() error {

	controller, err := warewulfconf.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	if controller.Nfs.Enabled {
		if err != nil {
			fmt.Println(err)
		}
		if controller.Warewulf.EnableHostOverlay {
			err = overlay.BuildHostOverlay()
			if err != nil {
				wwlog.Printf(wwlog.WARN, "host overlay could not be built: %s\n", err)
			}
		} else {
			wwlog.Printf(wwlog.INFO, "host overlays are disabled, did not modify exports")
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
