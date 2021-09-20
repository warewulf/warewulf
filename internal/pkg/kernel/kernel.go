package kernel

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/pkg/errors"

	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

func ParentDir() string {
	return path.Join(config.LocalStateDir, "provision/kernel")
}

func KernelImage(kernelName string) string {
	if kernelName == "" {
		wwlog.Printf(wwlog.ERROR, "Kernel Name is not defined\n")
		return ""
	}

	if !util.ValidString(kernelName, "^[a-zA-Z0-9-._]+$") {
		wwlog.Printf(wwlog.ERROR, "Runtime overlay name contains illegal characters: %s\n", kernelName)
		return ""
	}

	return path.Join(ParentDir(), kernelName, "vmlinuz")
}

func GetKernelVersion(kernelName string) string {
	if kernelName == "" {
		wwlog.Printf(wwlog.ERROR, "Kernel Name is not defined\n")
		return ""
	}
	kernelVersion, err := ioutil.ReadFile(path.Join(ParentDir(), kernelName, "version"))
	if err != nil {
		return ""
	}
	return string(kernelVersion)
}

func KmodsImage(kernelName string) string {
	if kernelName == "" {
		wwlog.Printf(wwlog.ERROR, "Kernel Name is not defined\n")
		return ""
	}

	if !util.ValidString(kernelName, "^[a-zA-Z0-9-._]+$") {
		wwlog.Printf(wwlog.ERROR, "Runtime overlay name contains illegal characters: %s\n", kernelName)
		return ""
	}

	return path.Join(ParentDir(), kernelName, "kmods.img")
}

func KernelVersion(kernelName string) string {
	if kernelName == "" {
		wwlog.Printf(wwlog.ERROR, "Kernel Name is not defined\n")
		return ""
	}

	if !util.ValidString(kernelName, "^[a-zA-Z0-9-._]+$") {
		wwlog.Printf(wwlog.ERROR, "Runtime overlay name contains illegal characters: %s\n", kernelName)
		return ""
	}

	return path.Join(ParentDir(), kernelName, "version")
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

func Build(kernelVersion string, root string, kernelName string) (string, error) {
	kernelImage := path.Join(root, "/boot/vmlinuz-"+kernelVersion)
	kernelDrivers := path.Join(root, "/lib/modules/"+kernelVersion)
	kernelDestination := KernelImage(kernelName)
	driversDestination := KmodsImage(kernelName)
	versionDestination := KernelVersion(kernelName)

	// Create the destination paths just in case it doesn't exist
	err := os.MkdirAll(path.Dir(kernelDestination), 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create kernal dest: %s", err)
	}

	err = os.MkdirAll(path.Dir(driversDestination), 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create driver dest: %s", err)
	}

	err = os.MkdirAll(path.Dir(versionDestination), 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create version dest: %s", err)
	}

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
		compressor, err := exec.LookPath("pigz")
		if err != nil {
			wwlog.Printf(wwlog.VERBOSE, "Could not locate PIGZ, using GZIP\n")
			compressor = "gzip"
		} else {
			wwlog.Printf(wwlog.VERBOSE, "Using PIGZ to compress the container: %s\n", compressor)
		}

		cmd := fmt.Sprintf("cd /; find .%s | cpio --quiet -o -H newc | %s -c > \"%s\"", kernelDrivers, compressor, driversDestination)

		wwlog.Printf(wwlog.DEBUG, "RUNNING: %s\n", cmd)
		err = exec.Command("/bin/sh", "-c", cmd).Run()
		if err != nil {
			return "", err
		}
	}

	wwlog.Printf(wwlog.VERBOSE, "Creating version file\n")
	file, err := os.Create(versionDestination)
	if err != nil {
		return "", errors.Wrap(err, "Failed to create version file")
	}
	defer file.Close()
	_, err = io.WriteString(file, kernelVersion)
	if err != nil {
		return "", errors.Wrap(err, "Could not write kernel version")
	}
	file.Sync()
	return "Done", nil
}

func DeleteKernel(name string) error {
	fullPath := path.Join(ParentDir(), name)

	wwlog.Printf(wwlog.VERBOSE, "Removing path: %s\n", fullPath)
	return os.RemoveAll(fullPath)
}
