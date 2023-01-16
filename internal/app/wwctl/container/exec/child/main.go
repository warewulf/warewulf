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
	containerPath := container.RootFsDir(containerName)

	mountFlags := uintptr(syscall.MS_PRIVATE | syscall.MS_REC)
	if unix.Access(containerPath, unix.W_OK) != nil {
		wwlog.Verbose("Path %s is readonly, mounting %s as ro", containerPath, containerName)
		mountFlags |= syscall.MS_RDONLY
	}
	err := syscall.Mount("", "/", "", mountFlags, "")
	if err != nil {
		return errors.Wrap(err, "failed to mount")
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

	os.Setenv("PS1", fmt.Sprintf("[%s] Warewulf> ", containerName))
	os.Setenv("PATH", "/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin")
	os.Setenv("HISTFILE", "/dev/null")

	err = syscall.Exec(args[1], args[1:], os.Environ())
	if err != nil {
		wwlog.Error("%s", err)
		os.Exit(1)
	}

	return nil
}
