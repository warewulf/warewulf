package vnfs

import (
	"context"
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/oci"
	"github.com/hpcng/warewulf/internal/pkg/vnfs"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"os"
	"path"
)


func BuildDocker(v vnfs.VnfsObject) {
	wwlog.Printf(wwlog.VERBOSE, "Building OCI Container: %s\n", v.Source())
	config := config.New()

	OciCacheDir := config.LocalStateDir + "/oci"
	VnfsHashDir := config.LocalStateDir + "/oci/vnfs"

	c, err := oci.NewCache(oci.OptSetCachePath(OciCacheDir))
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	wwlog.Printf(wwlog.VERBOSE, "Downloading OCI container layers\n")
	sourcePath, err := c.Pull(context.Background(), v.Source(), nil)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	hashDestination := VnfsHashDir + path.Base(sourcePath)

	name, err := os.Readlink(config.VnfsImage(v.NameClean()))
	if err == nil {
		if name == hashDestination {
			wwlog.Printf(wwlog.INFO, "%-35s: Skipping, VNFS is current\n", v.Name())
			return
		}
	}

	err = os.MkdirAll(VnfsHashDir, 0755)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
	err = os.MkdirAll(path.Dir(config.VnfsImage(v.NameClean())), 0755)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
	err = os.MkdirAll(path.Dir(config.VnfsChroot(v.NameClean())), 0755)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	wwlog.Printf(wwlog.VERBOSE, "Building bootable VNFS image\n")

	err = buildVnfs(sourcePath, hashDestination)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	wwlog.Printf(wwlog.VERBOSE, "Finalizing Build\n")

	// Setup links from OCI image to provision path
	_ = os.Remove(config.VnfsImage(v.NameClean()) + "-link")
	err = os.Symlink(hashDestination, config.VnfsImage(v.NameClean())+"-link")
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}
	err = os.Rename(config.VnfsImage(v.NameClean())+"-link", config.VnfsImage(v.NameClean()))
	if err != nil {
		os.Exit(1)
	}

	// Setup links from OCI rootfs to chroot path
	_ = os.Remove(config.VnfsChroot(v.NameClean()) + "-link")
	err = os.Symlink(sourcePath, config.VnfsChroot(v.NameClean())+"-link")
	if err != nil {
		os.Exit(1)
	}
	err = os.Rename(config.VnfsChroot(v.NameClean())+"-link", config.VnfsChroot(v.NameClean()))
	if err != nil {
		os.Exit(1)
	}


	wwlog.Printf(wwlog.INFO, "%-35s: Done\n", v.Name())

	return
}

