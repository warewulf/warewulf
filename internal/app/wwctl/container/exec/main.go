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
	// copy the files into the container at this stage, es in __child the
	// command syscall.Exec which replaces the __child process with the
	// exec command in the container. All the mounts, have to be done in
	// __child so that the used mounts don't propagate outside on the host
	// (see the CLONE attributes), but as for the cow copy option we need
	// to see if a file was modified after it was copied into the container
	// so do this here.
	// At first read out conf, the parse commandline, as copy files has the
	// same synatx as mount points
	mountPts := append(container.InitMountPnts(binds), conf.MountsContainer...)
	filesToCpy := getCopyFiles(mountPts)
	for _, cpyFile := range filesToCpy {
		if err := (cpyFile).copyToContainer(containerName); err != nil {
			wwlog.Warn("couldn't copy file: %s", err)
		}
	}
	wwlog.Verbose("Running contained command: %s", childArgs)
	retVal := childCommandFunc(cmd, childArgs)
	for _, cpyFile := range filesToCpy {
		if err := cpyFile.removeFromContainer(containerName); err != nil {
			wwlog.Warn("couldn't remove file: %s", err)
		}
	}
	return retVal
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

// file name and last modification time so we can remove the file if it wasn't modified
type cowFile struct {
	fileName string
	src      string
	modTime  time.Time
}

func (this *cowFile) copyToContainer(containerName string) error {
	containerDest := path.Join(container.RootFsDir(containerName), this.fileName)
	if _, err := os.Stat(path.Dir(containerDest)); err != nil {
		return fmt.Errorf("destination directory doesn't exist: %s", err)
	} else if _, err := os.Stat(containerDest); err == nil {
		return fmt.Errorf("file already exists in container: %s", this.fileName)
	} else if _, err := os.Stat(this.src); err != nil {
		return fmt.Errorf("source doesn't exist: %s", err)
	} else if err := util.CopyFile(this.src, containerDest); err != nil {
		return fmt.Errorf("couldn't copy files into container: %w", err)
	} else {
		// we can ignore error as the file was copied
		stat, _ := os.Stat(containerDest)
		this.modTime = stat.ModTime()
		return nil
	}
}

func (this *cowFile) removeFromContainer(containerName string) error {
	containerDest := path.Join(container.RootFsDir(containerName), this.fileName)
	if this.modTime.IsZero() {
		return fmt.Errorf("not previously copied: %s", this.fileName)
	} else if destStat, err := os.Stat(containerDest); err != nil {
		return fmt.Errorf("no longer present: %s", err)
	} else if destStat.ModTime() == this.modTime {
		return os.Remove(containerDest)
	} else {
		return fmt.Errorf("modified since copy: %s", this.fileName)
	}
}

/*
Check the objects we want to copy in, instead of mounting
*/
func getCopyFiles(binds []*warewulfconf.MountEntry) (copyObjects []*cowFile) {
	for _, bind := range binds {
		if bind.Cow {
			copyObjects = append(copyObjects, &cowFile{
				fileName: bind.Dest,
				src:      bind.Source,
			})
		}
	}
	return
}
