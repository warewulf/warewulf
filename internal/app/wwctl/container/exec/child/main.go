//go:build linux
// +build linux

package child

import (
	"fmt"
	"os"
	"path"
	"syscall"
	"time"

	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	if os.Getpid() != 1 {
		wwlog.Printf(wwlog.ERROR, "PID is not 1: %d\n", os.Getpid())
		os.Exit(1)
	}

	containerName := args[0]

	if !container.ValidSource(containerName) {
		wwlog.Printf(wwlog.ERROR, "Unknown Warewulf container: %s\n", containerName)
		os.Exit(1)
	}
	containerPath := container.RootFsDir(containerName)
	fileStat, _ := os.Stat(path.Join(containerPath, "/etc/passwd"))
	unixStat := fileStat.Sys().(*syscall.Stat_t)
	passwdTime := time.Unix(int64(unixStat.Ctim.Sec), int64(unixStat.Ctim.Nsec))
	fileStat, _ = os.Stat(path.Join(containerPath, "/etc/group"))
	unixStat = fileStat.Sys().(*syscall.Stat_t)
	groupTime := time.Unix(int64(unixStat.Ctim.Sec), int64(unixStat.Ctim.Nsec))
	wwlog.Printf(wwlog.DEBUG, "passwd: %v\n", passwdTime)
	wwlog.Printf(wwlog.DEBUG, "group: %v\n", groupTime)

	err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
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
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}
	fileStat, _ = os.Stat(path.Join(containerPath, "/etc/passwd"))
	unixStat = fileStat.Sys().(*syscall.Stat_t)
	if passwdTime.Before(time.Unix(int64(unixStat.Ctim.Sec), int64(unixStat.Ctim.Nsec))) {
		wwlog.Printf(wwlog.WARN, "/etc/passwd has been modified, maybe you want to run syncuser")
	}
	wwlog.Printf(wwlog.DEBUG, "passwd: %v\n", time.Unix(int64(unixStat.Ctim.Sec), int64(unixStat.Ctim.Nsec)))
	fileStat, _ = os.Stat(path.Join(containerPath, "/etc/group"))
	unixStat = fileStat.Sys().(*syscall.Stat_t)
	if groupTime.Before(time.Unix(int64(unixStat.Ctim.Sec), int64(unixStat.Ctim.Nsec))) {
		wwlog.Printf(wwlog.WARN, "/etc/group has been modified, maybe you want to run syncuser")
	}
	wwlog.Printf(wwlog.DEBUG, "group: %v\n", time.Unix(int64(unixStat.Ctim.Sec), int64(unixStat.Ctim.Nsec)))

	return nil
}
