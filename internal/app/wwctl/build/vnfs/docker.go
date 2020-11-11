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

const (
	OciCacheDir = config.LocalStateDir + "/oci"
	VnfsHashDir = config.LocalStateDir + "/oci/vnfs/hash/"
)

func BuildDocker(v vnfs.VnfsObject) {
	wwlog.Printf(wwlog.VERBOSE, "Building OCI Container: %s\n", v.Source())

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

	name, err := os.Readlink(v.Image())
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
	err = os.MkdirAll(path.Dir(v.Image()), 0755)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
	err = os.MkdirAll(path.Dir(v.Root()), 0755)
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

	_ = os.Remove(v.Image() + "-link")
	err = os.Symlink(hashDestination, v.Image()+"-link")
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}
	err = os.Rename(v.Image()+"-link", v.Image())

	err = buildLinks(v, sourcePath)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	wwlog.Printf(wwlog.INFO, "%-35s: Done\n", v.Name())

	return
}

