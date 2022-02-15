package nfs

import (
	"fmt"
	"os"

	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	return Configure(SetShow)
}

func Configure(show bool) error {

	controller, err := warewulfconf.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	if controller.Network == "" {
		wwlog.Printf(wwlog.ERROR, "Network must be defined in warewulf.conf to configure NFS\n")
		os.Exit(1)
	}
	if controller.Netmask == "" {
		wwlog.Printf(wwlog.ERROR, "Netmask must be defined in warewulf.conf to configure NFS\n")
		os.Exit(1)
	}

	if controller.Nfs.Enabled && !SetShow {
		// remove exports as templating may fail on existing files
		err := os.Remove("/etc/exports")
		if err != nil {
			fmt.Println(err)
		}
		err = overlay.BuildHostOverlay()
		if err != nil {
			wwlog.Printf(wwlog.WARN, "host overlay could not be built: %s\n", err)
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
