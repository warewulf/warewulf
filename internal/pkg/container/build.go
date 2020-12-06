package container

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"os"
	"os/exec"
	"path"
)

func Build(name string, buildForce bool) {

	rootfsPath := RootFsDir(name)
	imagePath := ImageFile(name)

	if ValidSource(name) == false {
		wwlog.Printf(wwlog.INFO, "%-35s: Skipping (bad path)\n", name)
		return
	}

	if buildForce == false {
		wwlog.Printf(wwlog.DEBUG, "Checking if there have been any updates to the VNFS directory\n")
		if util.PathIsNewer(rootfsPath, imagePath) {
			wwlog.Printf(wwlog.INFO, "%-35s: Skipping, VNFS is current\n", name)
			return
		}
	}

	wwlog.Printf(wwlog.DEBUG, "Making parent directory for: %s\n", name)
	err := os.MkdirAll(path.Dir(imagePath), 0755)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	wwlog.Printf(wwlog.DEBUG, "Making parent directory for: %s\n", rootfsPath)
	err = os.MkdirAll(path.Dir(rootfsPath), 0755)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	wwlog.Printf(wwlog.DEBUG, "Building VNFS image: '%s' -> '%s'\n", rootfsPath, imagePath)
	cmd := fmt.Sprintf("cd %s; find . | cpio --quiet -o -H newc | gzip -c > \"%s\"", rootfsPath, imagePath)
	err = exec.Command("/bin/sh", "-c", cmd).Run()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Failed building VNFS: %s\n", err)
		return
	}

	wwlog.Printf(wwlog.INFO, "%-35s: Done\n", name)

}
