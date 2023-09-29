package kernel

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"

	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

var (
	kernelSearchPaths = []string{
		// This is a printf format where the %s will be the kernel version
		"/boot/vmlinuz-linux%.s",
		"/boot/vmlinuz-%s",
		"/boot/vmlinuz-%s.gz",
		"/lib/modules/%s/vmlinuz",
		"/lib/modules/%s/vmlinuz.gz",
	}
)

func KernelImageTopDir() string {
	conf := warewulfconf.Get()
	return path.Join(conf.Paths.WWProvisiondir, "kernel")
}

func KernelImage(kernelName string) string {
	if kernelName == "" {
		wwlog.Error("Kernel Name is not defined")
		return ""
	}

	if !util.ValidString(kernelName, "^[a-zA-Z0-9-._]+$") {
		wwlog.Error("Runtime overlay name contains illegal characters: %s", kernelName)
		return ""
	}

	return path.Join(KernelImageTopDir(), kernelName, "vmlinuz")
}

func GetKernelVersion(kernelName string) string {
	if kernelName == "" {
		wwlog.Error("Kernel Name is not defined")
		return ""
	}
	kernelVersion, err := os.ReadFile(KernelVersionFile(kernelName))
	if err != nil {
		return ""
	}
	return string(kernelVersion)
}

func KmodsImage(kernelName string) string {
	if kernelName == "" {
		wwlog.Error("Kernel Name is not defined")
		return ""
	}

	if !util.ValidString(kernelName, "^[a-zA-Z0-9-._]+$") {
		wwlog.Error("Runtime overlay name contains illegal characters: %s", kernelName)
		return ""
	}

	return path.Join(KernelImageTopDir(), kernelName, "kmods.img")
}

func KernelVersionFile(kernelName string) string {
	if kernelName == "" {
		wwlog.Error("Kernel Name is not defined")
		return ""
	}

	if !util.ValidString(kernelName, "^[a-zA-Z0-9-._]+$") {
		wwlog.Error("Runtime overlay name contains illegal characters: %s", kernelName)
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

	wwlog.Debug("Searching for Kernel image directories: %s", KernelImageTopDir())

	kernels, err := os.ReadDir(KernelImageTopDir())
	if err != nil {
		return ret, err
	}

	for _, kernel := range kernels {
		wwlog.Verbose("Found Kernel: %s", kernel.Name())

		ret = append(ret, kernel.Name())

	}

	return ret, nil
}

/*
Triggers the kernel extraction and build of the modules for the given
kernel version. A name for this kernel and were to find has also to be
supplied
*/
func Build(kernelVersion, kernelName, root string) error {
	kernelDrivers := []string{path.Join("lib/modules/", kernelVersion, "*"), "lib/firmware/*"}
	kernelDestination := KernelImage(kernelName)
	driversDestination := KmodsImage(kernelName)
	versionDestination := KernelVersionFile(kernelName)
	var kernelSource string

	// Create the destination paths just in case it doesn't exist
	err := os.MkdirAll(path.Dir(kernelDestination), 0755)
	if err != nil {
		return errors.Wrap(err, "failed to create kernel dest")
	}

	err = os.MkdirAll(path.Dir(driversDestination), 0755)
	if err != nil {
		return errors.Wrap(err, "failed to create driver dest")
	}

	err = os.MkdirAll(path.Dir(versionDestination), 0755)
	if err != nil {
		return fmt.Errorf("failed to create version dest: %s", err)
	}

	for _, searchPath := range kernelSearchPaths {
		testPath := fmt.Sprintf(path.Join(root, searchPath), kernelVersion)
		wwlog.Verbose("Looking for kernel at: %s", testPath)
		if util.IsFile(testPath) {
			kernelSource = testPath
			break
		}
	}

	if kernelSource == "" {
		wwlog.Error("Could not locate kernel image")
		return errors.New("could not locate kernel image")
	} else {
		wwlog.Info("Found kernel at: %s", kernelSource)
	}

	wwlog.Verbose("Setting up Kernel")
	if _, err := os.Stat(kernelSource); err == nil {
		kernel, err := os.Open(kernelSource)
		if err != nil {
			return errors.Wrap(err, "could not open kernel")
		}
		defer kernel.Close()

		gzipreader, err := gzip.NewReader(kernel)
		if err == nil {
			defer gzipreader.Close()

			writer, err := os.Create(kernelDestination)
			if err != nil {
				return errors.Wrap(err, "could not decompress kernel")
			}
			defer writer.Close()

			_, err = io.Copy(writer, gzipreader)
			if err != nil {
				return errors.Wrap(err, "could not write decompressed kernel")
			}

		} else {

			err := util.CopyFile(kernelSource, kernelDestination)
			if err != nil {
				return errors.Wrap(err, "could not copy kernel")
			}
		}

	}

	name := kernelName + " drivers"
	wwlog.Verbose("Creating image for %s: %s", name, root)

	err = util.BuildFsImage(
		name,
		root,
		driversDestination,
		kernelDrivers,
		[]string{},
		// ignore cross-device files
		true,
		"newc",
		// dereference symbolic links
		"-L")

	if err != nil {
		return err
	}

	wwlog.Verbose("Creating version file")
	file, err := os.Create(versionDestination)
	if err != nil {
		return errors.Wrap(err, "Failed to create version file")
	}
	defer file.Close()
	_, err = io.WriteString(file, kernelVersion)
	if err != nil {
		return errors.Wrap(err, "Could not write kernel version")
	}
	err = file.Sync()
	if err != nil {
		return errors.Wrap(err, "Could not sync kernel version")
	}
	return nil
}

func DeleteKernel(name string) error {
	fullPath := path.Join(KernelImageTopDir(), name)

	wwlog.Verbose("Removing path: %s", fullPath)
	return os.RemoveAll(fullPath)
}

func FindKernelVersion(root string) (string, error) {
	for _, searchPath := range kernelSearchPaths {
		testPattern := fmt.Sprintf(path.Join(root, searchPath), `*`)
		wwlog.Verbose("Looking for kernel version with pattern at: %s", testPattern)
		potentialKernel, _ := filepath.Glob(testPattern)
		if len(potentialKernel) == 0 {
			continue
		}
		for _, foundKernel := range potentialKernel {
			wwlog.Verbose("Parsing out kernel version for %s", foundKernel)
			re := regexp.MustCompile(fmt.Sprintf(path.Join(root, searchPath), `([\w\d-\.]*)`))
			version := re.FindAllStringSubmatch(foundKernel, -1)
			if version == nil {
				return "", fmt.Errorf("could not parse kernel version")
			}
			wwlog.Verbose("found kernel version %s", version)
			return version[0][1], nil

		}

	}
	return "", fmt.Errorf("could not find kernel version")

}
