package vnfs

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/vnfs"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"os"
	"path"
)

func BuildContainerdir(v vnfs.VnfsObject) {
	config := config.New()

	if _, err := os.Stat(v.Source()); err != nil {
		wwlog.Printf(wwlog.INFO, "%-35s: Skipping (bad path)\n", v.Name())
		return
	}

	wwlog.Printf(wwlog.DEBUG, "Checking if there have been any updates to the VNFS directory\n")
	if util.PathIsNewer(v.Source(), config.VnfsImage(v.NameClean())) {
		if buildForce == false {
			wwlog.Printf(wwlog.INFO, "%-35s: Skipping, VNFS is current\n", v.Name())
			return
		}
	}

	wwlog.Printf(wwlog.DEBUG, "Making the directory: %s\n", path.Dir(config.VnfsImage(v.NameClean())))
	err := os.MkdirAll(path.Dir(config.VnfsImage(v.NameClean())), 0755)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	wwlog.Printf(wwlog.DEBUG, "Making the directory: %s\n", path.Dir(config.VnfsChroot(v.NameClean())))
	err = os.MkdirAll(path.Dir(config.VnfsChroot(v.NameClean())), 0755)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	wwlog.Printf(wwlog.DEBUG, "Building VNFS image: '%s' -> '%s'\n", v.Source(), config.VnfsImage(v.NameClean()))
	err = buildVnfs(v.Source(), config.VnfsImage(v.NameClean()))
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	// Setup links from OCI rootfs to chroot path
	_ = os.Remove(config.VnfsChroot(v.NameClean()) + "-link")
	err = os.Symlink(v.Source(), config.VnfsChroot(v.NameClean())+"-link")
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not create symlink for Chroot: %s\n", err)
		os.Exit(1)
	}
	err = os.Rename(config.VnfsChroot(v.NameClean())+"-link", config.VnfsChroot(v.NameClean()))
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not rename link: %s\n", err)
		os.Exit(1)
	}

	wwlog.Printf(wwlog.INFO, "%-35s: Done\n", v.Name())

}
