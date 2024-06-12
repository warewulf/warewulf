//go:build linux
// +build linux

package exec

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"syscall"
	"time"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/container"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func runChildCmd(cmd *cobra.Command, args []string) error {
	child := exec.Command("/proc/self/exe", args...)
	child.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	child.Stdin = cmd.InOrStdin()
	child.Stdout = cmd.OutOrStdout()
	child.Stderr = cmd.ErrOrStderr()
	return child.Run()
}

var childCommandFunc = runChildCmd

// Fork a child process with a new PID space
func runContainedCmd(cmd *cobra.Command, containerName string, args []string) (err error) {
	wwlog.Debug("runContainedCmd:args: %v", args)

	conf := warewulfconf.Get()

	runDir := container.RunDir(containerName)
	if err := os.Mkdir(runDir, 0750); err != nil {
		if _, existerr := os.Stat(runDir); !os.IsNotExist(existerr) {
			return errors.New("run directory already exists: another container command may already be running")
		} else {
			return fmt.Errorf("unable to create run directory: %w", err)
		}
	}
	defer func() {
		if err := os.RemoveAll(runDir); err != nil {
			wwlog.Error("error removing run directory: %w", err)
		}
	}()

	logStr := fmt.Sprint(wwlog.GetLogLevel())

	childArgs := []string{"--warewulfconf", conf.GetWarewulfConf(), "--loglevel", logStr, "container", "exec", "__child"}
	childArgs = append(childArgs, containerName)
	for _, b := range binds {
		childArgs = append(childArgs, "--bind", b)
	}
	if nodeName != "" {
		childArgs = append(childArgs, "--node", nodeName)
	}
	childArgs = append(childArgs, args...)
	wwlog.Verbose("Running contained command: %s", childArgs)
	return childCommandFunc(cmd, childArgs)
}

func CobraRunE(cmd *cobra.Command, args []string) error {
	wwlog.Debug("CobraRunE:args: %v", args)

	containerName := args[0]
	wwlog.Debug("CobraRunE:containerName: %v", containerName)
	if !container.ValidSource(containerName) {
		wwlog.Error("Unknown Warewulf container: %s", containerName)
		os.Exit(1)
	}
	os.Setenv("WW_CONTAINER_SHELL", containerName)

	containerPath := container.RootFsDir(containerName)

	beforePasswdTime := getTime(path.Join(containerPath, "/etc/passwd"))
	wwlog.Debug("passwdTime: %v", beforePasswdTime)
	beforeGroupTime := getTime(path.Join(containerPath, "/etc/group"))
	wwlog.Debug("groupTime: %v", beforeGroupTime)

	err := runContainedCmd(cmd, containerName, args[1:])
	if err != nil {
		wwlog.Error("Failed executing container command: %s", err)
		os.Exit(1)
	}

	if util.IsFile(path.Join(containerPath, "/etc/warewulf/container_exit.sh")) {
		wwlog.Verbose("Found clean script: /etc/warewulf/container_exit.sh")
		err = runContainedCmd(cmd, containerName, []string{"/bin/sh", "/etc/warewulf/container_exit.sh"})
		if err != nil {
			wwlog.Error("Failed executing exit script: %s", err)
			os.Exit(1)
		}
	}

	userdbChanged := false
	if !beforePasswdTime.IsZero() {
		afterPasswdTime := getTime(path.Join(containerPath, "/etc/passwd"))
		wwlog.Debug("passwdTime: %v", afterPasswdTime)
		if beforePasswdTime.Before(afterPasswdTime) {
			if !SyncUser {
				wwlog.Warn("/etc/passwd has been modified, maybe you want to run syncuser")
			}
			userdbChanged = true
		}
	}
	if !beforeGroupTime.IsZero() {
		afterGroupTime := getTime(path.Join(containerPath, "/etc/group"))
		wwlog.Debug("groupTime: %v", afterGroupTime)
		if beforeGroupTime.Before(afterGroupTime) {
			if !SyncUser {
				wwlog.Warn("/etc/group has been modified, maybe you want to run syncuser")
			}
			userdbChanged = true
		}
	}
	if userdbChanged && SyncUser {
		err = container.SyncUids(containerName, false)
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

func getTime(path string) time.Time {
	if fileStat, err := os.Stat(path); err != nil {
		return time.Time{}
	} else {
		unixStat := fileStat.Sys().(*syscall.Stat_t)
		return time.Unix(int64(unixStat.Ctim.Sec), int64(unixStat.Ctim.Nsec))
	}
}

func SetBinds(myBinds []string) {
	binds = append(binds, myBinds...)
}

func SetNode(myNode string) {
	nodeName = myNode
}
