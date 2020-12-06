package vnfs

import (
	"context"
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/oci"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"os"
	"path"
)

func BuildDocker(vnfs VnfsObject, buildForce bool) {
	wwlog.Printf(wwlog.VERBOSE, "Building OCI Container: %s\n", vnfs.Source)

	OciCacheDir := config.LocalStateDir + "/oci"
	VnfsHashDir := config.LocalStateDir + "/oci/vnfs"

	c, err := oci.NewCache(oci.OptSetCachePath(OciCacheDir))
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	wwlog.Printf(wwlog.VERBOSE, "Downloading OCI container layers\n")
	sourcePath, err := c.Pull(context.Background(), vnfs.Source, nil)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	hashDestination := VnfsHashDir + path.Base(sourcePath)

	name, err := os.Readlink(vnfs.Image)
	if err == nil {
		if name == hashDestination && buildForce == false {
			wwlog.Printf(wwlog.INFO, "%-35s: Skipping, VNFS is current\n", vnfs.Name)
			return
		}
	}

	err = os.MkdirAll(VnfsHashDir, 0755)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	wwlog.Printf(wwlog.DEBUG, "Making parent directory for: %s\n", vnfs.Image)
	err = os.MkdirAll(path.Dir(vnfs.Image), 0755)
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

	wwlog.Printf(wwlog.VERBOSE, "Building bootable VNFS image\n")

	err = buildVnfs(sourcePath, hashDestination)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	wwlog.Printf(wwlog.VERBOSE, "Finalizing Build\n")

	// Setup links from OCI image to provision path
	_ = os.Remove(vnfs.Image + "-link")
	err = os.Symlink(hashDestination, vnfs.Image+"-link")
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not create symlink for Image: %s\n", err)
		os.Exit(1)
	}
	err = os.Rename(vnfs.Image+"-link", vnfs.Image)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not rename link: %s\n", err)
		os.Exit(1)
	}

	// Setup links from OCI rootfs to chroot path
	_ = os.Remove(vnfs.Chroot + "-link")
	err = os.Symlink(sourcePath, vnfs.Chroot+"-link")
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not create symlink for Chroot: %s\n", err)
		os.Exit(1)
	}
	err = os.Rename(vnfs.Chroot+"-link", vnfs.Chroot)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not rename link: %s\n", err)
		os.Exit(1)
	}

	wwlog.Printf(wwlog.INFO, "%-35s: Done\n", vnfs.Source)

	return
}
