package container

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/pkg/errors"

	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

func Build(name string, buildForce bool) error {

	rootfsPath := RootFsDir(name)
	imagePath := ImageFile(name)

	if !ValidSource(name) {
		return errors.New("Container does not exist")
	}

	if !buildForce {
		wwlog.Printf(wwlog.DEBUG, "Checking if there have been any updates to the VNFS directory\n")
		if util.PathIsNewer(rootfsPath, imagePath) {
			wwlog.Printf(wwlog.INFO, "Skipping (VNFS is current)\n")
			return nil
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

	compressor, err := exec.LookPath("pigz")
	if err != nil {
		wwlog.Printf(wwlog.VERBOSE, "Could not locate PIGZ, using GZIP\n")
		compressor = "gzip"
	} else {
		wwlog.Printf(wwlog.VERBOSE, "Using PIGZ to compress the container: %s\n", compressor)
	}
	var cmd string
	_, err = os.Stat(path.Join(rootfsPath, "./etc/warewulf/excludes"))
	if os.IsNotExist(err) {
		wwlog.Printf(wwlog.DEBUG, "Building VNFS image: '%s' -> '%s'\n", rootfsPath, imagePath)
		cmd = fmt.Sprintf("cd %s; find . -xdev -xautofs | cpio --quiet -o -H newc | %s -c > \"%s\"", rootfsPath, compressor, imagePath)
	} else {
		wwlog.Printf(wwlog.DEBUG, "Building VNFS image with excludes: '%s' -> '%s'\n", rootfsPath, imagePath)
		cmd = fmt.Sprintf("cd %s; find . -xdev -xautofs | grep -v -f ./etc/warewulf/excludes | cpio --quiet -o -H newc | %s -c > \"%s\"", rootfsPath, compressor, imagePath)
	}
	wwlog.Printf(wwlog.DEBUG, "RUNNING: %s\n", cmd)
	err = exec.Command("/bin/sh", "-c", cmd).Run()
	if err != nil {
		return errors.New("Failed building VNFS")
	}

	return nil
}
