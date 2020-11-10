package main

import (
	"context"
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/oci"
	"github.com/hpcng/warewulf/internal/pkg/vnfs"
	"log"
	"os"
	"path"
	"sync"
)

const (
	OciCacheDir = config.LocalStateDir + "/oci"
	VnfsHashDir = config.LocalStateDir + "/oci/vnfs/hash/"
)

func vnfsOciBuild(OciPath string, wg *sync.WaitGroup) {
	v := vnfs.New(OciPath)

	//	vnfsDestination := fmt.Sprintf("%s/provision/vnfs/%s.img.gz", LocalStateDir, path.Base(OciPath))
	defer wg.Done()

	log.Printf("Building OCI Container: %s\n", OciPath)

	c, err := oci.NewCache(oci.OptSetCachePath(OciCacheDir))
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}
	log.Printf("Downloading OCI container layers\n")
	sourcePath, err := c.Pull(context.Background(), OciPath, nil)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	hashDestination := VnfsHashDir + path.Base(sourcePath)

	name, err := os.Readlink(v.Image())
	if err == nil {
		if name == hashDestination {
			log.Printf("Container already built, no update available\n")
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

	log.Printf("Building bootable VNFS image\n")

	err = buildVnfs(sourcePath, hashDestination)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	log.Printf("Finalizing Build\n")

	_ = os.Remove(v.Image() + "-link")
	err = os.Symlink(hashDestination, v.Image()+"-link")
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}
	err = os.Rename(v.Image()+"-link", v.Image())

	err = buildLinks(v, sourcePath)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	log.Printf("Completed building VNFS: %s\n", path.Base(OciPath))

	return
}
