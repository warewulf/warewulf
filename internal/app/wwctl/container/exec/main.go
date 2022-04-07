//go:build linux
// +build linux

package exec

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"syscall"
	"time"

	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func runContainedCmd(args []string) error {
	wwlog.Printf(wwlog.VERBOSE, "Running contained command: %s\n", args[1:])
	c := exec.Command("/proc/self/exe", append([]string{"container", "exec", "__child"}, args...)...)

	c.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return nil
}

func CobraRunE(cmd *cobra.Command, args []string) error {

	containerName := args[0]
	var allargs []string

	if !container.ValidSource(containerName) {
		wwlog.Printf(wwlog.ERROR, "Unknown Warewulf container: %s\n", containerName)
		os.Exit(1)
	}

	for _, b := range binds {
		allargs = append(allargs, "--bind", b)
	}
	allargs = append(allargs, args...)
	containerPath := container.RootFsDir(containerName)

	fileStat, _ := os.Stat(path.Join(containerPath, "/etc/passwd"))
	unixStat := fileStat.Sys().(*syscall.Stat_t)
	passwdTime := time.Unix(int64(unixStat.Ctim.Sec), int64(unixStat.Ctim.Nsec))
	fileStat, _ = os.Stat(path.Join(containerPath, "/etc/group"))
	unixStat = fileStat.Sys().(*syscall.Stat_t)
	groupTime := time.Unix(int64(unixStat.Ctim.Sec), int64(unixStat.Ctim.Nsec))
	wwlog.Printf(wwlog.DEBUG, "passwd: %v\n", passwdTime)
	wwlog.Printf(wwlog.DEBUG, "group: %v\n", groupTime)

	err := runContainedCmd(allargs)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Failed executing container command: %s\n", err)
		os.Exit(1)
	}

	if util.IsFile(path.Join(container.RootFsDir(allargs[0]), "/etc/warewulf/container_exit.sh")) {
		wwlog.Printf(wwlog.VERBOSE, "Found clean script: /etc/warewulf/container_exit.sh\n")
		err = runContainedCmd([]string{allargs[0], "/bin/sh", "/etc/warewulf/container_exit.sh"})
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Failed executing exit script: %s\n", err)
			os.Exit(1)
		}
	}
	fileStat, _ = os.Stat(path.Join(containerPath, "/etc/passwd"))
	unixStat = fileStat.Sys().(*syscall.Stat_t)
	syncuids := false
	if passwdTime.Before(time.Unix(int64(unixStat.Ctim.Sec), int64(unixStat.Ctim.Nsec))) {
		if NoSyncUser {
			wwlog.Printf(wwlog.WARN, "/etc/passwd has been modified, maybe you want to run syncuser\n")
		}
		syncuids = true
	}
	wwlog.Printf(wwlog.DEBUG, "passwd: %v\n", time.Unix(int64(unixStat.Ctim.Sec), int64(unixStat.Ctim.Nsec)))
	fileStat, _ = os.Stat(path.Join(containerPath, "/etc/group"))
	unixStat = fileStat.Sys().(*syscall.Stat_t)
	if groupTime.Before(time.Unix(int64(unixStat.Ctim.Sec), int64(unixStat.Ctim.Nsec))) {
		if NoSyncUser {
			wwlog.Printf(wwlog.WARN, "/etc/group has been modified, maybe you want to run syncuser\n")
		}
		syncuids = true
	}
	wwlog.Printf(wwlog.DEBUG, "group: %v\n", time.Unix(int64(unixStat.Ctim.Sec), int64(unixStat.Ctim.Nsec)))
	if syncuids && !NoSyncUser {
		err = container.SyncUids(containerName, true)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Error in user sync, fix error and run 'syncuser' manually, but trying to build container: %s\n", err)
		}
	}

	fmt.Printf("Rebuilding container...\n")
	err = container.Build(containerName, false)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not build container %s: %s\n", containerName, err)
		os.Exit(1)
	}

	return nil
}
