package kernel

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/pkg/errors"

	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

var (
	kernelSearchPaths = []string{
		// This is a printf format where the %s will be the kernel version
		"/boot/vmlinuz-%s",
		"/boot/vmlinuz-%s.gz",
		"/lib/mmodules/%s/vmlinuz",
		"/lib/mmodules/%s/vmlinuz.gz",
	}
)

func KernelImageTopDir() string {
	return path.Join(buildconfig.WWPROVISIONDIR, "kernel")
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

	return path.Join(KernelImageTopDir(), kernelName, "vmlinuz")
}

func GetKernelVersion(kernelName string) string {
	if kernelName == "" {
		wwlog.Printf(wwlog.ERROR, "Kernel Name is not defined\n")
		return ""
	}
	kernelVersion, err := ioutil.ReadFile(KernelVersionFile(kernelName))
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

	return path.Join(KernelImageTopDir(), kernelName, "kmods.img")
}

func KernelVersionFile(kernelName string) string {
	if kernelName == "" {
		wwlog.Printf(wwlog.ERROR, "Kernel Name is not defined\n")
		return ""
	}

	if !util.ValidString(kernelName, "^[a-zA-Z0-9-._]+$") {
		wwlog.Printf(wwlog.ERROR, "Runtime overlay name contains illegal characters: %s\n", kernelName)
		return ""
	}

	return path.Join(KernelImageTopDir(), kernelName, "version")
}

func ListKernels() ([]string, error) {
	var ret []string

	err := os.MkdirAll(KernelImageTopDir(), 0755)
	if err != nil {
		return ret, errors.New("Could not create Kernel parent directory: " + KernelImageTopDir())
	}

	wwlog.Printf(wwlog.DEBUG, "Searching for Kernel image directories: %s\n", KernelImageTopDir())

	kernels, err := ioutil.ReadDir(KernelImageTopDir())
	if err != nil {
		return ret, err
	}

	for _, kernel := range kernels {
		wwlog.Printf(wwlog.VERBOSE, "Found Kernel: %s\n", kernel.Name())

		ret = append(ret, kernel.Name())

	}

	return ret, nil
}

func Build(kernelVersion, kernelName, root string) (string, error) {
	kernelDrivers := path.Join(root, "/lib/modules/"+kernelVersion)
	kernelDriversRelative := path.Join("/lib/modules/" + kernelVersion)
	kernelDestination := KernelImage(kernelName)
	driversDestination := KmodsImage(kernelName)
	versionDestination := KernelVersionFile(kernelName)
	var kernelSource string

	// Create the destination paths just in case it doesn't exist
	err := os.MkdirAll(path.Dir(kernelDestination), 0755)
	if err != nil {
		return "", errors.Wrap(err, "failed to create kernel dest")
	}

	err = os.MkdirAll(path.Dir(driversDestination), 0755)
	if err != nil {
		return "", errors.Wrap(err, "failed to create driver dest")
	}

	err = os.MkdirAll(path.Dir(versionDestination), 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create version dest: %s", err)
	}

	for _, path := range kernelSearchPaths {
		testPath := fmt.Sprintf(path, kernelVersion)
		wwlog.Printf(wwlog.VERBOSE, "Looking for kernel at: %s\n", testPath)
		if util.IsFile(testPath) {
			kernelSource = testPath
			break
		}
	}

	if kernelSource == "" {
		wwlog.Printf(wwlog.ERROR, "Could not locate kernel image\n")
		return "", errors.New("could not locate kernel image")
	} else {
		wwlog.Printf(wwlog.INFO, "Found kernel at: %s\n", kernelSource)
	}

	if !util.IsDir(kernelDrivers) {
		return "", errors.New("Could not locate kernel drivers")
	}

	wwlog.Printf(wwlog.VERBOSE, "Setting up Kernel\n")
	if _, err := os.Stat(kernelSource); err == nil {
		kernel, err := os.Open(kernelSource)
		if err != nil {
			return "", errors.Wrap(err, "could not open kernel")
		}
		defer kernel.Close()

		gzipreader, err := gzip.NewReader(kernel)
		if err == nil {
			defer gzipreader.Close()

			writer, err := os.Create(kernelDestination)
			if err != nil {
				return "", errors.Wrap(err, "could not decompress kernel")
			}
			defer writer.Close()

			_, err = io.Copy(writer, gzipreader)
			if err != nil {
				return "", errors.Wrap(err, "could not write decompressed kernel")
			}

		} else {

			err := util.CopyFile(kernelSource, kernelDestination)
			if err != nil {
				return "", errors.Wrap(err, "could not copy kernel")
			}
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

		cmd := fmt.Sprintf("cd %s; find .%s | cpio --quiet -o -H newc | %s -c > \"%s\"", root, kernelDriversRelative, compressor, driversDestination)

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
	err = file.Sync()
	if err != nil {
		return "", errors.Wrap(err, "Could not sync kernel version")
	}
	return "Done", nil
}

func DeleteKernel(name string) error {
	fullPath := path.Join(KernelImageTopDir(), name)

	wwlog.Printf(wwlog.VERBOSE, "Removing path: %s\n", fullPath)
	return os.RemoveAll(fullPath)
}
