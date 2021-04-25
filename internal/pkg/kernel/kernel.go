package kernel

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/errors"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

func ParentDir() string {
	return path.Join(config.LocalStateDir, "provision/kernel")
}

func KernelImage(kernelVersion string) string {
	if kernelVersion == "" {
		wwlog.Printf(wwlog.ERROR, "Kernel Version is not defined\n")
		return ""
	}

	if !util.ValidString(kernelVersion, "^[a-zA-Z0-9-._]+$") {
		wwlog.Printf(wwlog.ERROR, "Runtime overlay name contains illegal characters: %s\n", kernelVersion)
		return ""
	}

	return path.Join(ParentDir(), kernelVersion, "vmlinuz")
}

func KmodsImage(kernelVersion string) string {
	if kernelVersion == "" {
		wwlog.Printf(wwlog.ERROR, "Kernel Version is not defined\n")
		return ""
	}

	if !util.ValidString(kernelVersion, "^[a-zA-Z0-9-._]+$") {
		wwlog.Printf(wwlog.ERROR, "Runtime overlay name contains illegal characters: %s\n", kernelVersion)
		return ""
	}

	return path.Join(ParentDir(), kernelVersion, "kmods.img")
}

func ListKernels() ([]string, error) {
	var ret []string

	err := os.MkdirAll(ParentDir(), 0755)
	if err != nil {
		return ret, errors.New("Could not create Kernel parent directory: " + ParentDir())
	}

	wwlog.Printf(wwlog.DEBUG, "Searching for Kernel image directories: %s\n", ParentDir())

	kernels, err := ioutil.ReadDir(ParentDir())
	if err != nil {
		return ret, err
	}

	for _, kernel := range kernels {
		wwlog.Printf(wwlog.VERBOSE, "Found Kernel: %s\n", kernel.Name())

		ret = append(ret, kernel.Name())

	}

	return ret, nil
}

func Build(kernelVersion string, root string) (string, error) {

	kernelImage := path.Join(root, "/boot/vmlinuz-"+kernelVersion)
	kernelDrivers := path.Join(root, "/lib/modules/"+kernelVersion)
	kernelDestination := KernelImage(kernelVersion)
	driversDestination := KmodsImage(kernelVersion)

	// Create the destination paths just in case it doesn't exist
	os.MkdirAll(path.Dir(kernelDestination), 0755)
	os.MkdirAll(path.Dir(driversDestination), 0755)

	if !util.IsFile(kernelImage) {
		return "", errors.New("Could not locate kernel image")
	}

	if !util.IsDir(kernelDrivers) {
		return "", errors.New("Could not locate kernel drivers")
	}

	wwlog.Printf(wwlog.VERBOSE, "Setting up Kernel\n")
	if _, err := os.Stat(kernelImage); err == nil {
		err := util.CopyFile(kernelImage, kernelDestination)
		if err != nil {
			return "", err
		}
	}

	wwlog.Printf(wwlog.VERBOSE, "Building Kernel driver image\n")
	if _, err := os.Stat(kernelDrivers); err == nil {
		cmd := fmt.Sprintf("cd /; find .%s | cpio --quiet -o -H newc | gzip -c > \"%s\"", kernelDrivers, driversDestination)
		// use pigz if available
		err := exec.Command("/bin/sh", "-c", "command -v pigz").Run()
		if err == nil {
			cmd = fmt.Sprintf("cd /; find .%s | cpio --quiet -o -H newc | pigz -c > \"%s\"", kernelDrivers, driversDestination)
		}
		wwlog.Printf(wwlog.DEBUG, "RUNNING: %s\n", cmd)
		err = exec.Command("/bin/sh", "-c", cmd).Run()
		if err != nil {
			return "", err
		}
	}

	return "Done", nil
}

func DeleteKernel(name string) error {
	fullPath := path.Join(ParentDir(), name)

	wwlog.Printf(wwlog.VERBOSE, "Removing path: %s\n", fullPath)
	return os.RemoveAll(fullPath)
}
