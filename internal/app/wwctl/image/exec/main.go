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

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/image"
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
func runContainedCmd(cmd *cobra.Command, imageName string, args []string) (err error) {
	wwlog.Debug("runContainedCmd:args: %v", args)

	conf := warewulfconf.Get()

	runDir := image.RunDir(imageName)
	if err := os.Mkdir(runDir, 0750); err != nil {
		if _, existerr := os.Stat(runDir); !os.IsNotExist(existerr) {
			return fmt.Errorf("run directory already exists: another image command may already be running (otherwise, remove %s)", runDir)
		} else {
			return fmt.Errorf("unable to create run directory: %w", err)
		}
	}
	defer func() {
		if err := os.RemoveAll(runDir); err != nil {
			wwlog.Error("error removing run directory: %s", err)
		}
	}()

	logStr := fmt.Sprint(wwlog.GetLogLevel())
	childArgs := []string{"--warewulfconf", conf.GetWarewulfConf(), "--loglevel", logStr, "image", "exec", "__child"}
	childArgs = append(childArgs, imageName)
	for _, b := range binds {
		childArgs = append(childArgs, "--bind", b)
	}
	if nodeName != "" {
		childArgs = append(childArgs, "--node", nodeName)
	}
	childArgs = append(childArgs, "--")
	childArgs = append(childArgs, args...)
	// copy the files into the image at this stage, es in __child the
	// command syscall.Exec which replaces the __child process with the
	// exec command in the image. All the mounts, have to be done in
	// __child so that the used mounts don't propagate outside on the host
	// (see the CLONE attributes), but as for the copy option we need
	// to see if a file was modified after it was copied into the image
	// so do this here.
	// At first read out conf, the parse commandline, as copy files has the
	// same synatx as mount points
	mountPts := append(image.InitMountPnts(binds), conf.MountsImage...)
	filesToCpy := getCopyFiles(mountPts)
	for _, cpyFile := range filesToCpy {
		if err := (cpyFile).copyToImage(imageName); err != nil {
			wwlog.Warn("couldn't copy file: %s", err)
		}
	}
	wwlog.Verbose("Running contained command: %s", childArgs)
	retVal := childCommandFunc(cmd, childArgs)
	for _, cpyFile := range filesToCpy {
		if cpyFile.shouldRemoveFromImage(imageName) {
			if err := cpyFile.removeFromImage(imageName); err != nil {
				wwlog.Warn("couldn't remove file: %s", err)
			}
		}
	}
	return retVal
}

func CobraRunE(cmd *cobra.Command, args []string) error {
	wwlog.Debug("CobraRunE:args: %v", args)

	imageName := args[0]
	wwlog.Debug("CobraRunE:imageName: %v", imageName)
	if !image.ValidSource(imageName) {
		return fmt.Errorf("unknown Warewulf image: %s", imageName)
	}
	os.Setenv("WW_CONTAINER_SHELL", imageName)
	os.Setenv("WW_IMAGE_SHELL", imageName)

	imagePath := image.RootFsDir(imageName)

	beforePasswdTime := getTime(path.Join(imagePath, "/etc/passwd"))
	wwlog.Debug("passwdTime: %v", beforePasswdTime)
	beforeGroupTime := getTime(path.Join(imagePath, "/etc/group"))
	wwlog.Debug("groupTime: %v", beforeGroupTime)

	err := runContainedCmd(cmd, imageName, args[1:])
	if err != nil {
		return fmt.Errorf("command returned an error: %v: %s", args[1:], err)
	}

	for _, exitScript := range []string{"/etc/warewulf/image_exit.sh", "/etc/warewulf/container_exit.sh"} {
		if util.IsFile(path.Join(imagePath, exitScript)) {
			wwlog.Verbose("Found exit script: %s", exitScript)
			err = runContainedCmd(cmd, imageName, []string{"/bin/sh", exitScript})
			if err != nil {
				return fmt.Errorf("exit script returned an error: %v: %s", exitScript, err)
			}
			break
		}
	}

	userdbChanged := false
	if !beforePasswdTime.IsZero() {
		afterPasswdTime := getTime(path.Join(imagePath, "/etc/passwd"))
		wwlog.Debug("passwdTime: %v", afterPasswdTime)
		if beforePasswdTime.Before(afterPasswdTime) {
			userdbChanged = true
		}
	}
	if !beforeGroupTime.IsZero() {
		afterGroupTime := getTime(path.Join(imagePath, "/etc/group"))
		wwlog.Debug("groupTime: %v", afterGroupTime)
		if beforeGroupTime.Before(afterGroupTime) {
			userdbChanged = true
		}
	}
	if SyncUser {
		if userdbChanged {
			if err = image.Syncuser(imageName, false); err != nil {
				wwlog.Error("syncuser error: %s", err)
			}
		} else {
			wwlog.Info("Skipping syncuser (passwd and group files not changed)")
		}
	}

	if Build {
		err = image.Build(imageName, false)
		if err != nil {
			return fmt.Errorf("could not build image: %s: %s", imageName, err)
		}
	}
	return nil
}

func getTime(path string) time.Time {
	if fileStat, err := os.Stat(path); err != nil {
		return time.Time{}
	} else {
		unixStat := fileStat.Sys().(*syscall.Stat_t)
		return time.Unix(int64(unixStat.Mtim.Sec), int64(unixStat.Mtim.Nsec))
	}
}

func SetBinds(myBinds []string) {
	binds = append(binds, myBinds...)
}

func SetNode(myNode string) {
	nodeName = myNode
}

// file name and last modification time so we can remove the file if it wasn't modified
type copyFile struct {
	fileName string
	src      string
	modTime  time.Time
}

func (cf *copyFile) imageDest(imageName string) string {
	return path.Join(image.RootFsDir(imageName), cf.fileName)
}

func (cf *copyFile) copyToImage(imageName string) error {
	imageDest := cf.imageDest(imageName)
	if _, err := os.Stat(path.Dir(imageDest)); err != nil {
		return err
	} else if _, err := os.Stat(imageDest); err == nil {
		return err
	} else if _, err := os.Stat(cf.src); err != nil {
		return err
	} else if err := util.CopyFile(cf.src, imageDest); err != nil {
		return err
	} else if stat, err := os.Stat(imageDest); err != nil {
		return err
	} else {
		cf.modTime = stat.ModTime()
		return nil
	}
}

func (cf *copyFile) shouldRemoveFromImage(imageName string) bool {
	imageDest := cf.imageDest(imageName)
	if cf.modTime.IsZero() {
		wwlog.Debug("file was not previously copied: %s", cf.fileName)
		return false
	} else if destStat, err := os.Stat(imageDest); err != nil {
		wwlog.Verbose("file is no longer present: %s (%s)", cf.fileName, err)
		return false
	} else if destStat.ModTime() == cf.modTime {
		wwlog.Verbose("don't remove modified file:", cf.fileName)
		return false
	} else {
		return true
	}
}

func (cf *copyFile) removeFromImage(imageName string) error {
	imageDest := cf.imageDest(imageName)
	return os.Remove(imageDest)
}

/*
Check the objects we want to copy in, instead of mounting
*/
func getCopyFiles(binds []*warewulfconf.MountEntry) (copyObjects []*copyFile) {
	for _, bind := range binds {
		if bind.Copy() {
			copyObjects = append(copyObjects, &copyFile{
				fileName: bind.Dest,
				src:      bind.Source,
			})
		}
	}
	return
}
