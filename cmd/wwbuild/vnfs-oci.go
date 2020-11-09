package main

import (
	"context"
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/oci"
	"log"
	"os"
	"path"
	"sync"
)

func vnfsOciBuild(OciPath string, wg *sync.WaitGroup) {
	vnfsDestination := fmt.Sprintf("%s/provision/vnfs/%s.img.gz", LocalStateDir, path.Base(OciPath))
	defer wg.Done()

	log.Printf("Building OCI Container: %s\n", OciPath)
	c, err := oci.NewCache(oci.OptSetCachePath("/var/warewulf/oci"))
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}
	sourcePath, err := c.Pull(context.Background(), OciPath, nil)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	ociDestination := fmt.Sprintf("%s/oci/vnfs/hash/%s", LocalStateDir, path.Base(sourcePath))

	name, err := os.Readlink(vnfsDestination)
	if err == nil {
		if name == ociDestination {
			log.Printf("Container already built, no changes from upstream\n")
			return
		}
	}

	err = os.MkdirAll(path.Dir(ociDestination), 0755)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
	err = os.MkdirAll(path.Dir(vnfsDestination), 0755)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
	err = os.MkdirAll(LocalStateDir+"/oci/vnfs/name", 0755)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	err = buildVnfs(sourcePath, ociDestination)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	err = os.Symlink(ociDestination, vnfsDestination+"-link")
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	err = os.Symlink(sourcePath, LocalStateDir+"/oci/vnfs/name/"+path.Base(OciPath))
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	err = os.Rename(vnfsDestination+"-link", vnfsDestination)

	log.Printf("Completed building VNFS: %s\n", path.Base(OciPath))

	return
}
