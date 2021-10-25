package nfs

import (
	"fmt"
	"os"

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

	if !SetShow {
		fstab, err := os.OpenFile("/var/warewulf/overlays/system/default/etc/fstab", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}
		defer fstab.Close()

		fmt.Fprintf(fstab, "# This file was written by Warewulf (wwctl configure nfs)\n")

		fmt.Fprintf(fstab, "rootfs / tmpfs defaults 0 0\n")
		fmt.Fprintf(fstab, "devpts /dev/pts devpts gid=5,mode=620 0 0\n")
		fmt.Fprintf(fstab, "tmpfs /run/shm tmpfs defaults 0 0\n")
		fmt.Fprintf(fstab, "sysfs /sys sysfs defaults 0 0\n")
		fmt.Fprintf(fstab, "proc /proc proc defaults 0 0\n")

		if controller.Nfs.Enabled {
			exports, err := os.OpenFile("/etc/exports", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
			defer exports.Close()

			fmt.Fprintf(exports, "# This file was written by Warewulf (wwctl configure nfs)\n")

			for _, export := range controller.Nfs.Exports {
				fmt.Fprintf(exports, "%s %s/%s(sync)\n", export, controller.Network, controller.Netmask)
				fmt.Fprintf(fstab, "%s:%s %s nfs defaults 0 0\n", controller.Ipaddr, export, export)
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
	} else {
		fmt.Printf("/etc/exports:\n")
		for _, export := range controller.Nfs.Exports {
			fmt.Printf("%s %s/%s\n", export, controller.Network, controller.Netmask)
		}
		fmt.Printf("\n")
		fmt.Printf("SYSTEM OVERLAY: default/etc/fstab:\n")
		for _, export := range controller.Nfs.Exports {
			fmt.Printf("%s:%s %s nfs defaults 0 0\n", controller.Ipaddr, export, export)
		}

	}

	return nil
}
