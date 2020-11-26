package vnfs

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"os"
	"path"
)

func BuildContainerdir(vnfs VnfsObject, buildForce bool) {

	if _, err := os.Stat(vnfs.Source); err != nil {
		wwlog.Printf(wwlog.INFO, "%-35s: Skipping (bad path)\n", vnfs.Source)
		return
	}

	wwlog.Printf(wwlog.DEBUG, "Checking if there have been any updates to the VNFS directory\n")
	if util.PathIsNewer(vnfs.Source, vnfs.Image) {
		if buildForce == false {
			wwlog.Printf(wwlog.INFO, "%-35s: Skipping, VNFS is current\n", vnfs.Name)
			return
		}
	}

	wwlog.Printf(wwlog.DEBUG, "Making parent directory for: %s\n", vnfs.Image)
	err := os.MkdirAll(path.Dir(vnfs.Image), 0755)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	wwlog.Printf(wwlog.DEBUG, "Making parent directory for: %s\n", vnfs.Chroot)
	err = os.MkdirAll(path.Dir(vnfs.Chroot), 0755)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	wwlog.Printf(wwlog.DEBUG, "Building VNFS image: '%s' -> '%s'\n", vnfs.Source, vnfs.Image)
	err = buildVnfs(vnfs.Source, vnfs.Image)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	// Setup links from OCI rootfs to chroot path
	_ = os.Remove(vnfs.Chroot + "-link")
	err = os.Symlink(vnfs.Source, vnfs.Chroot+"-link")
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not create symlink for Chroot: %s\n", err)
		os.Exit(1)
	}
	err = os.Rename(vnfs.Chroot + "-link", vnfs.Chroot)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not rename link: %s\n", err)
		os.Exit(1)
	}

	wwlog.Printf(wwlog.INFO, "%-35s: Done\n", vnfs.Name)

}
