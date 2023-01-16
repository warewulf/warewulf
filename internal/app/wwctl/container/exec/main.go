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

/*
fork off a process with a new PID space
*/
func runContainedCmd(args []string) error {
	logStr := fmt.Sprint(wwlog.GetLogLevel())
	wwlog.Verbose("Running contained command: %s", args[1:])
	c := exec.Command("/proc/self/exe", append([]string{"--loglevel", logStr, "container", "exec", "__child"}, args...)...)

	c.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		fmt.Printf("Command exited non-zero, not rebuilding/updating VNFS image\n")
		os.Exit(0)
	}
	return nil
}

func CobraRunE(cmd *cobra.Command, args []string) error {

	containerName := args[0]
	os.Setenv("WW_CONTAINER_SHELL", containerName)

	var allargs []string

	if !container.ValidSource(containerName) {
		wwlog.Error("Unknown Warewulf container: %s", containerName)
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
	wwlog.Debug("passwd: %v", passwdTime)
	wwlog.Debug("group: %v", groupTime)

	err := runContainedCmd(allargs)
	if err != nil {
		wwlog.Error("Failed executing container command: %s", err)
		os.Exit(1)
	}

	if util.IsFile(path.Join(container.RootFsDir(allargs[0]), "/etc/warewulf/container_exit.sh")) {
		wwlog.Verbose("Found clean script: /etc/warewulf/container_exit.sh")
		err = runContainedCmd([]string{allargs[0], "/bin/sh", "/etc/warewulf/container_exit.sh"})
		if err != nil {
			wwlog.Error("Failed executing exit script: %s", err)
			os.Exit(1)
		}
	}
	fileStat, _ = os.Stat(path.Join(containerPath, "/etc/passwd"))
	unixStat = fileStat.Sys().(*syscall.Stat_t)
	syncuids := false
	if passwdTime.Before(time.Unix(int64(unixStat.Ctim.Sec), int64(unixStat.Ctim.Nsec))) {
		if !SyncUser {
			wwlog.Warn("/etc/passwd has been modified, maybe you want to run syncuser")
		}
		syncuids = true
	}
	wwlog.Debug("passwd: %v", time.Unix(int64(unixStat.Ctim.Sec), int64(unixStat.Ctim.Nsec)))
	fileStat, _ = os.Stat(path.Join(containerPath, "/etc/group"))
	unixStat = fileStat.Sys().(*syscall.Stat_t)
	if groupTime.Before(time.Unix(int64(unixStat.Ctim.Sec), int64(unixStat.Ctim.Nsec))) {
		if !SyncUser {
			wwlog.Warn("/etc/group has been modified, maybe you want to run syncuser")
		}
		syncuids = true
	}
	wwlog.Debug("group: %v", time.Unix(int64(unixStat.Ctim.Sec), int64(unixStat.Ctim.Nsec)))
	if syncuids && SyncUser {
		err = container.SyncUids(containerName, true)
		if err != nil {
			wwlog.Error("Error in user sync, fix error and run 'syncuser' manually, but trying to build container: %s", err)
		}
	}

	fmt.Printf("Rebuilding container...\n")
	err = container.Build(containerName, false)
	if err != nil {
		wwlog.Error("Could not build container %s: %s", containerName, err)
		os.Exit(1)
	}

	return nil
}
func SetBinds(myBinds []string) {
	binds = append(binds, myBinds...)
}
