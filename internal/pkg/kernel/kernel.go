package kernel

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"github.com/hashicorp/go-version"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

var (
	kernelSearchPaths = []string{
		// This is a printf format where the %s will be the kernel version
		"/boot/Image-%s", // this is the aarch64 for SUSE, vmlinux which is also present won't boot
		"/boot/vmlinuz-linux%s",
		"/boot/vmlinuz-%s",
		"/boot/vmlinuz-%s.gz",
		"/lib/modules/%s/vmlinuz",
		"/lib/modules/%s/vmlinuz.gz",
	}
	kernelDrivers = []string{
		"lib/modules/%s/*",
		"lib/firmware/*",
		"lib/modprobe.d",
		"lib/modules-load.d"}
	// kenrel naming convention <base kernel version>-<ABI number>.<upload number>-<flavour>
	kernelVersionRegex = `(\d+\.\d+\.\d+)-((\d+\.*){1,})`
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
	kernelDestination := KernelImage(kernelName)
	driversDestination := KmodsImage(kernelName)
	versionDestination := KernelVersionFile(kernelName)

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

	kernelSource, kernelVersFound, err := FindKernel(root)
	if err != nil {
		return err
	} else if kernelVersFound != kernelVersion {
		return fmt.Errorf("requested %s and found kernel version %s differ", kernelVersion, kernelVersFound)
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
	var kernelDriversSpecific []string
	for _, kPath := range kernelDrivers {
		if strings.Contains(kPath, "%s") {
			kernelDriversSpecific = append(kernelDriversSpecific, fmt.Sprintf(kPath, kernelVersion))
		} else {
			kernelDriversSpecific = append(kernelDriversSpecific, kPath)
		}
	}
	wwlog.Debug("kernelDriversSpecific: %v", kernelDriversSpecific)
	wwlog.Verbose("Creating image for %s: %s", name, root)
	err = util.BuildFsImage(
		name,
		root,
		driversDestination,
		kernelDriversSpecific,
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

/*
Searches for kernel under a given path. First return result is the
full path, second the version and an error if the kernel couldn't be found.
*/
type kernel struct {
	version string
	path    string
}

func filter(val string, filters []func(string) (string, error)) (string, error) {
	for _, ft := range filters {
		newVal, err := ft(val)
		if err != nil {
			return val, err
		}
		val = newVal
	}
	return val, nil
}

func nonDebugKernel(val string) (string, error) {
	if strings.HasSuffix(val, "+debug") {
		return val, fmt.Errorf("%s is debug kernel, skipped", val)
	}
	return val, nil
}

func nonSemaVer(val string) (string, error) {
	// need to extract version info
	verRegx := regexp.MustCompile(kernelVersionRegex)
	verRe := verRegx.FindAllStringSubmatch(val, -1)
	// only if at the least the following pattern is matched <xx.xx.xx>-<xx[.xx.xx]>
	if len(verRe) > 0 && len(verRe[0]) > 2 {
		// verRe[0][1] -> <xx.xx.xx>
		// verRe[0][2] -> <xx[.xx.xx]>
		verStr := strings.TrimSuffix(fmt.Sprintf("%s-%s", verRe[0][1], verRe[0][2]), ".")
		_, err := version.NewVersion(verStr)
		if err != nil {
			return val, fmt.Errorf("semantic incompatible version detected, version string: %s, err: %s", verStr, err)
		}
		return verStr, nil
	}
	return val, fmt.Errorf("unable to extract version info from %s", val)
}

func FindKernel(root string) (string, string, error) {
	wwlog.Debug("root: %s", root)
	for _, searchPath := range kernelSearchPaths {
		testPattern := fmt.Sprintf(path.Join(root, searchPath), `*`)
		wwlog.Debug("Looking for kernel version with pattern at: %s", testPattern)
		potentialKernel, _ := filepath.Glob(testPattern)
		if len(potentialKernel) == 0 {
			continue
		}

		verMap := make(map[*version.Version]*kernel, len(potentialKernel))
		for _, foundKernel := range potentialKernel {
			wwlog.Debug("Parsing out kernel version for %s", foundKernel)
			re := regexp.MustCompile(fmt.Sprintf(path.Join(root, searchPath), `([\w\d-\.+]*)`))
			kernelVer := re.FindAllStringSubmatch(foundKernel, -1)
			if kernelVer == nil {
				break
			}
			// kernelVerStr is like 5.14.0-427.18.1.el9_4.x86_64
			kernelVerStr := strings.TrimSuffix(kernelVer[0][1], ".gz")

			newVal, err := filter(kernelVerStr, []func(string) (string, error){nonDebugKernel, nonSemaVer})
			if err != nil {
				wwlog.Verbose("While filtering kernel version for %s, having error: %s", kernelVerStr, err)
				continue
			}
			ver, _ := version.NewVersion(newVal)
			verMap[ver] = &kernel{
				version: kernelVerStr,
				path:    foundKernel,
			}
		}

		if len(verMap) > 0 {
			var keys []*version.Version
			for k := range verMap {
				keys = append(keys, k)
			}
			sort.Sort(sort.Reverse(version.Collection(keys)))
			return verMap[keys[0]].path, verMap[keys[0]].version, nil
		}
	}
	return "", "", fmt.Errorf("could not find kernel version")

}
