//go:build linux
// +build linux

package child

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	if os.Getpid() != 1 {
		wwlog.Error("PID is not 1: %d", os.Getpid())
		os.Exit(1)
	}

	containerName := args[0]

	if !container.ValidSource(containerName) {
		wwlog.Error("Unknown Warewulf container: %s", containerName)
		os.Exit(1)
	}
	var err error
	tempDir := ""
	// check for valid mount points
	lowerObjects := checkMountPoints(containerName, binds)
	if len(lowerObjects) != 0 {
		// need to create a overlay, where the lower layer contains
		// the missing mount points
		tempDir, err = os.MkdirTemp(buildconfig.TMPDIR(), "overlay")
		if err == nil {
			wwlog.Verbose("for ephermal mount use tempdir %s", tempDir)
			// ignore errors as we are doomed if a tmp dir couldn't be written
			_ = os.Mkdir(path.Join(tempDir, "work"), os.ModePerm)
			_ = os.Mkdir(path.Join(tempDir, "lower"), os.ModePerm)
			for _, obj := range lowerObjects {
				newFile := ""
				if !strings.HasSuffix(obj, "/") {
					newFile = filepath.Base(obj)
					obj = filepath.Dir(obj)
				}
				err = os.MkdirAll(filepath.Join(tempDir, "lower", obj), os.ModePerm)
				if err != nil {
					wwlog.Warn("couldn't create directory for mounts: %s", err)
				}
				if newFile != "" {
					desc, err := os.Create(filepath.Join(tempDir, "lower", obj, newFile))
					if err != nil {
						wwlog.Warn("couldn't create directory for mounts: %s", err)
					}
					defer desc.Close()
				}

			}
		} else {
			wwlog.Warn("couldn't create temp dir for overlay", err)
			lowerObjects = []string{}
		}
		/*
			\TODO check why defer isn't called at exit
			defer func() {
				fmt.Println("defer called")
				_ = os.RemoveAll(tempDir)
			}()
		*/
	}
	containerPath := container.RootFsDir(containerName)
	err = syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	if err != nil {
		return errors.Wrap(err, "failed to mount")
	}
	if len(lowerObjects) != 0 {
		options := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s",
			path.Join(tempDir, "lower"), containerPath, path.Join(tempDir, "work"))
		wwlog.Debug("overlay options: %s", options)
		err = syscall.Mount("overlay", containerPath, "overlay", 0, options)
		if err != nil {
			wwlog.Warn(fmt.Sprintf("Couldn't create overlay for ephermal mount points: %s", err))
		}
	}
	ps1Str := fmt.Sprintf("[%s] Warewulf> ", containerName)
	if !util.IsWriteAble(containerPath) {
		wwlog.Verbose("mounting %s ro", containerPath)
		ps1Str = fmt.Sprintf("[%s] (ro) Warewulf> ", containerName)
		err = syscall.Mount(containerPath, containerPath, "", syscall.MS_BIND, "")
		if err != nil {
			return errors.Wrap(err, "failed to prepare bind mount")
		}
		err = syscall.Mount(containerPath, containerPath, "", syscall.MS_REMOUNT|syscall.MS_RDONLY|syscall.MS_BIND, "")
		if err != nil {
			return errors.Wrap(err, "failed to remount ro")
		}
	}

	err = syscall.Mount("/dev", path.Join(containerPath, "/dev"), "", syscall.MS_BIND, "")
	if err != nil {
		return errors.Wrap(err, "failed to mount /dev")
	}

	for _, b := range binds {
		var source string
		var dest string

		bind := util.SplitValidPaths(b, ":")
		source = bind[0]

		if len(bind) == 1 {
			dest = source
		} else {
			dest = bind[1]
		}

		err := syscall.Mount(source, path.Join(containerPath, dest), "", syscall.MS_BIND, "")
		if err != nil {
			fmt.Printf("BIND ERROR: %s\n", err)
			wwlog.Warn("Couldn't mount %s", source)
		}
		wwlog.Verbose("mounted from host to container: %s:%s", source, dest)
	}

	err = syscall.Chroot(containerPath)
	if err != nil {
		return errors.Wrap(err, "failed to chroot")
	}

	err = os.Chdir("/")
	if err != nil {
		return errors.Wrap(err, "failed to chdir")
	}

	err = syscall.Mount("/proc", "/proc", "proc", 0, "")
	if err != nil {
		return errors.Wrap(err, "failed to mount proc")
	}

	os.Setenv("PS1", ps1Str)
	os.Setenv("PATH", "/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin")
	os.Setenv("HISTFILE", "/dev/null")

	err = syscall.Exec(args[1], args[1:], os.Environ())
	fmt.Println("After exec")
	_ = os.RemoveAll(tempDir)
	if err != nil {
		wwlog.Error("%s", err)

		os.Exit(1)
	}

	return nil
}

/*
Check if the bind mount points exists in the given container. Returns
the invalid mount points. Directories always have '/' as suffix
*/
func checkMountPoints(containerName string, binds []string) (overlayObjects []string) {
	wwlog.Debug("Checking if container %s has paths: %v", containerName, binds)
	overlayObjects = []string{}
	for _, b := range binds {
		var source string
		var dest string

		bind := util.SplitValidPaths(b, ":")
		source = bind[0]

		if len(bind) == 1 {
			dest = source
		} else {
			dest = bind[1]
		}
		err, _ := os.Stat(source)
		if err == nil {
			// no need to create a mount location if source doesn't exist
			continue
		}
		if _, err := os.Stat(path.Join(container.RootFsDir(containerName), dest)); err != nil {
			if os.IsNotExist(err) {
				if util.IsDir(dest) && !strings.HasSuffix(dest, "/") {
					dest += "/"
				}
				overlayObjects = append(overlayObjects, source)
				wwlog.Debug("Container %s, needs following path: %s", containerName, dest)
			}
		}
	}
	return overlayObjects
}
