//go:build linux
// +build linux

package child

import (
	"fmt"
	"os"
	"path"
	"syscall"

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
	// check for valid mount points
	containerPath := container.RootFsDir(containerName)
	err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	if err != nil {
		return errors.Wrap(err, "failed to mount")
	}
	ps1Str := fmt.Sprintf("[%s] Warewulf> ", containerName)
	if !util.IsWriteAble(containerPath) {
		wwlog.Verbose("mounting %s ro", containerPath)
		ps1Str = fmt.Sprintf("[%s] (ro) Warewulf> ", containerName)
		err = syscall.Mount(containerPath, containerPath, "", syscall.MS_BIND, "")
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to prepare bind mount"))
		}
		err = syscall.Mount(containerPath, containerPath, "", syscall.MS_REMOUNT|syscall.MS_RDONLY|syscall.MS_BIND, "")
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to remount ro"))
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
			os.Exit(1)
		}
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
	if err != nil {
		wwlog.Error("%s", err)
		os.Exit(1)
	}

	return nil
}
func checkMountPoints(containerName string, binds []string) (overlayObjects []string) {
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
		if err != nil {
			// no need to create a mount location if source doesn't exist
			continue
		}
		err, stat := os.Stat(path.Join(container.RootFsDir(containerName),dest))
		if err != nil {
			if 
		}
	}
}
