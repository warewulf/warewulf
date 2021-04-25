package container

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/hpcng/warewulf/internal/pkg/errors"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

func Build(name string, buildForce bool) error {

	rootfsPath := RootFsDir(name)
	imagePath := ImageFile(name)

	if ValidSource(name) == false {
		return errors.New("Container does not exist")
	}

	if buildForce == false {
		wwlog.Printf(wwlog.DEBUG, "Checking if there have been any updates to the VNFS directory\n")
		if util.PathIsNewer(rootfsPath, imagePath) {

			return errors.New("Skipping (VNFS is current)")
		}
	}

	wwlog.Printf(wwlog.DEBUG, "Making parent directory for: %s\n", name)
	err := os.MkdirAll(path.Dir(imagePath), 0755)
	if err != nil {
		return errors.New("Failed creating directory")
	}

	wwlog.Printf(wwlog.DEBUG, "Making parent directory for: %s\n", rootfsPath)
	err = os.MkdirAll(path.Dir(rootfsPath), 0755)
	if err != nil {
		return errors.New("Failed creating directory")
	}

	wwlog.Printf(wwlog.DEBUG, "Building VNFS image: '%s' -> '%s'\n", rootfsPath, imagePath)
	cmd := fmt.Sprintf("cd %s; find . | cpio --quiet -o -H newc | gzip -c > \"%s\"", rootfsPath, imagePath)
	// use pigz if available
	err = exec.Command("/bin/sh", "-c", "command -v pigz").Run()
	if err == nil {
		cmd = fmt.Sprintf("cd %s; find . | cpio --quiet -o -H newc | pigz -c > \"%s\"", rootfsPath, imagePath)
	}
	wwlog.Printf(wwlog.DEBUG, "RUNNING: %s\n", cmd)
	err = exec.Command("/bin/sh", "-c", cmd).Run()
	if err != nil {
		return errors.New("Failed building VNFS")
	}

	return nil
}
