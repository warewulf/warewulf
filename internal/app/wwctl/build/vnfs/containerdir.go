package vnfs

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/vnfs"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"os"
	"path"
)

func BuildContainerdir(v vnfs.VnfsObject) {

	if _, err := os.Stat(v.Source()); err != nil {
		wwlog.Printf(wwlog.INFO, "%-35s: Skipping (bad path)\n", v.Name())
		return
	}

	wwlog.Printf(wwlog.DEBUG, "Checking if there have been any updates to the VNFS directory\n")
	if util.PathIsNewer(v.Source(), v.Image()) {
		if buildForce == false {
			wwlog.Printf(wwlog.INFO, "%-35s: Skipping, VNFS is current\n", v.Name())
			return
		}
	}

	wwlog.Printf(wwlog.DEBUG, "Making the directory: %s\n", path.Dir(v.Image()))
	err := os.MkdirAll(path.Dir(v.Image()), 0755)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	wwlog.Printf(wwlog.DEBUG, "Building VNFS image: '%s' -> '%s'\n", v.Source(), v.Image())
	err = buildVnfs(v.Source(), v.Image())
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	wwlog.Printf(wwlog.DEBUG, "Building links for Warewulf access to chroot\n")
	err = buildLinks(v, v.Source())
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	wwlog.Printf(wwlog.INFO, "%-35s: Done\n", v.Name())

}
