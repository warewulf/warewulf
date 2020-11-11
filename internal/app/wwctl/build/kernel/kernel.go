package kernel

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/assets"
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"os"
	"os/exec"
	"path"
)

const kernelProvisionPath = "/provision/kernel/"

func Build(nodeList []assets.NodeInfo, force bool) error {
	set := make(map[string]int)

	wwlog.Printf(wwlog.INFO, "Importing Kernels:\n")
	wwlog.SetIndent(4)

	for _, node := range nodeList {
		if node.KernelVersion != "" {
			set[node.KernelVersion] ++
			wwlog.Printf(wwlog.DEBUG, "Node '%s' has KernelVersion '%s'\n", node.Fqdn, node.KernelVersion)
		}
	}

	for kernelVersion := range set {
		kernelImage := "/boot/vmlinuz-"+kernelVersion
		kernelDrivers := "/lib/modules/"+kernelVersion
		kernelDestination := path.Join(config.LocalStateDir, kernelProvisionPath, "vmlinuz-"+kernelVersion)
		driversDestination := path.Join(config.LocalStateDir, kernelProvisionPath, "kmods-"+kernelVersion+".img")


		// Create the kernel destination path just in case it doesn't exist
		os.MkdirAll(path.Join(config.LocalStateDir, kernelProvisionPath), 0755)

		if _, err := os.Stat(kernelImage); err == nil {
			if util.PathIsNewer(kernelImage, kernelDestination) && force == false {
				wwlog.Printf(wwlog.INFO, "%-35s: Skipping, kernel is current\n", "vmlinuz-"+kernelVersion)

			} else {
				err := util.CopyFile(kernelImage, kernelDestination)
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "Failed copying kernel image: %s\n", err)
					continue
				}
				wwlog.Printf(wwlog.INFO, "%-35s: Done\n", "vmlinuz-"+kernelVersion)

			}
		}

		if _, err := os.Stat(kernelDrivers); err == nil {
			if util.PathIsNewer(kernelDrivers, driversDestination) && force == false {
				wwlog.Printf(wwlog.INFO, "%-35s: Skipping, kernel Drivers are current\n", "kmods-"+kernelVersion+".img")

			} else {
				cmd := fmt.Sprintf("cd /; find .%s | cpio --quiet -o -H newc -F \"%s\"", kernelDrivers, driversDestination)
				err := exec.Command("/bin/sh", "-c", cmd).Run()
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "Could not generate kernel driver overlay: %s\n", err)
					continue
				}
				wwlog.Printf(wwlog.INFO, "%-35s: Done\n", "kmods-"+kernelVersion+".img")
			}
		}

	}

	wwlog.SetIndent(0)

	return nil
}