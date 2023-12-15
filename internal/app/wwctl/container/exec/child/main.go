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

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/container"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

const exitEval = `$(VALU="$?" ; if [ $VALU == 0 ]; then echo write; else echo discard; fi)`
const msgStr = `Container image is rebuilt depending on the exit status of the last command.

Run "true" or "false" to enforce or abort image rebuild.`

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	if os.Getpid() != 1 {
		wwlog.Error("PID is not 1: %d", os.Getpid())
		os.Exit(1)
	}

	containerName := args[0]

	if !container.ValidSource(containerName) {
		wwlog.Error("Unknown Warewulf container: %s", containerName)
		os.Exit(1)
	}
	conf := warewulfconf.Get()
	runDir := container.RunDir(containerName)
	if _, err := os.Stat(runDir); os.IsNotExist(err) {
		return errors.Wrap(err, "container run directory does not exist")
	}
	mountPts := conf.MountsContainer
	mountPts = append(container.InitMountPnts(binds), mountPts...)
	// check for valid mount points
	lowerObjects := checkMountPoints(containerName, mountPts)
	// need to create a overlay, where the lower layer contains
	// the missing mount points
	wwlog.Verbose("for ephermal mount use tempdir %s", runDir)
	if err = os.Mkdir(path.Join(runDir, "work"), os.ModePerm); err != nil {
		return err
	}
	if err = os.Mkdir(path.Join(runDir, "lower"), os.ModePerm); err != nil {
		return err
	}
	if err = os.Mkdir(path.Join(runDir, "nodeoverlay"), os.ModePerm); err != nil {
		return err
	}
	// handle all lower object, have some extra logic if the object is a file
	for _, obj := range lowerObjects {
		newFile := ""
		if !strings.HasSuffix(obj, "/") {
			newFile = filepath.Base(obj)
			obj = filepath.Dir(obj)
		}
		err = os.Mkdir(filepath.Join(runDir, "lower", obj), os.ModePerm)
		if err != nil {
			wwlog.Warn("couldn't create directory for mounts: %s", err)
		}
		if newFile != "" {
			desc, err := os.Create(filepath.Join(runDir, "lower", obj, newFile))
			if err != nil {
				wwlog.Warn("couldn't create directory for mounts: %s", err)
			}
			defer desc.Close()
		}
	}
	containerPath := container.RootFsDir(containerName)
	// running in a private PID space, so also make / private, so that nothing gets out from here
	err = syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	if err != nil {
		return errors.Wrap(err, "failed to mount")
	}
	ps1Str := fmt.Sprintf("[%s|%s] Warewulf> ", containerName, exitEval)
	wwlog.Info(msgStr)
	if len(lowerObjects) != 0 && nodename == "" {
		options := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s",
			path.Join(runDir, "lower"), containerPath, path.Join(runDir, "work"))
		wwlog.Debug("overlay options: %s", options)
		err = syscall.Mount("overlay", containerPath, "overlay", 0, options)
		if err != nil {
			wwlog.Warn("Couldn't create overlay for ephermal mount points: %s", err)
		}
	} else if nodename != "" {
		nodeDB, err := node.New()
		if err != nil {
			wwlog.Error("Could not open node configuration: %s", err)
			os.Exit(1)
		}

		nodes, err := nodeDB.FindAllNodes()
		if err != nil {
			wwlog.Error("Could not get node list: %s", err)
			os.Exit(1)
		}
		nodes = node.FilterByName(nodes, []string{nodename})
		if len(nodes) != 1 {
			wwlog.Error("No single node idendified with %s", nodename)
			os.Exit(1)
		}
		overlays := nodes[0].SystemOverlay
		overlays = append(overlays, nodes[0].RuntimeOverlay...)
		err = overlay.BuildOverlayIndir(nodes[0], overlays, path.Join(runDir, "nodeoverlay"))
		if err != nil {
			wwlog.Error("Could not build overlay: %s", err)
			os.Exit(1)
		}
		options := fmt.Sprintf("lowerdir=%s:%s:%s",
			path.Join(runDir, "lower"), containerPath, path.Join(runDir, "nodeoverlay"))
		wwlog.Debug("overlay options: %s", options)
		err = syscall.Mount("overlay", containerPath, "overlay", 0, options)
		if err != nil {
			wwlog.Warn(fmt.Sprintf("Couldn't create overlay for node render overlay: %s", err))
			os.Exit(1)
		}
		ps1Str = fmt.Sprintf("[%s|ro|%s] Warewulf> ", containerName, nodename)
	}
	if !util.IsWriteAble(containerPath) && nodename == "" {
		wwlog.Verbose("mounting %s ro", containerPath)
		ps1Str = fmt.Sprintf("[%s|ro] Warewulf> ", containerName)
		err = syscall.Mount(containerPath, containerPath, "", syscall.MS_BIND, "")
		if err != nil {
			return errors.Wrap(err, "failed to prepare bind mount")
		}
		err = syscall.Mount(containerPath, containerPath, "", syscall.MS_REMOUNT|syscall.MS_RDONLY|syscall.MS_BIND, "")
		if err != nil {
			return errors.Wrap(err, "failed to remount ro")
		}
	}

	for _, mntPnt := range mountPts {
		if mntPnt.Copy {
			continue
		}
		wwlog.Debug("bind mounting: %s -> %s", mntPnt.Source, path.Join(containerPath, mntPnt.Dest))
		err = syscall.Mount(mntPnt.Source, path.Join(containerPath, mntPnt.Dest), "", syscall.MS_BIND, "")
		if err != nil {
			wwlog.Warn("Couldn't mount %s to %s: %s", mntPnt.Source, mntPnt.Dest, err)
		} else if mntPnt.ReadOnly {
			err = syscall.Mount(mntPnt.Source, path.Join(containerPath, mntPnt.Dest), "", syscall.MS_REMOUNT|syscall.MS_RDONLY|syscall.MS_BIND, "")
			if err != nil {
				wwlog.Warn("failed to following mount readonly: %s", mntPnt.Source)
			} else {
				wwlog.Verbose("mounted readonly from host to container: %s:%s", mntPnt.Source, mntPnt.Dest)
			}
		} else {
			wwlog.Verbose("mounted from host to container: %s:%s", mntPnt.Source, mntPnt.Dest)
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

	if err := syscall.Mount("devtmpfs", "/dev", "devtmpfs", 0, ""); err != nil {
		return errors.Wrap(err, "failed to mount /dev")
	}
	if err := syscall.Mount("sysfs", "/sys", "sysfs", 0, ""); err != nil {
		return errors.Wrap(err, "failed to mount /sys")
	}
	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		return errors.Wrap(err, "failed to mount /proc")
	}
	if err := syscall.Mount("tmpfs", "/run", "tmpfs", 0, ""); err != nil {
		return errors.Wrap(err, "failed to mount /run")
	}

	os.Setenv("PS1", ps1Str)
	os.Setenv("PATH", "/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin")
	os.Setenv("HISTFILE", "/dev/null")

	wwlog.Debug("Exec: %s %s", args[1], args[1:])
	return syscall.Exec(args[1], args[1:], os.Environ())
}

/*
Check if the bind mount points exists in the given container. Returns
the invalid mount points. Directories always have '/' as suffix
*/
func checkMountPoints(containerName string, binds []*warewulfconf.MountEntry) (overlayObjects []string) {
	overlayObjects = []string{}
	for _, b := range binds {
		if b.Copy {
			continue
		}
		_, err := os.Stat(b.Source)
		if err != nil {
			wwlog.Debug("Couldn't stat %s create no mount point in container", b.Source)
			continue
		}
		wwlog.Debug("Checking in container for %s", path.Join(container.RootFsDir(containerName), b.Dest))
		if _, err = os.Stat(path.Join(container.RootFsDir(containerName), b.Dest)); err != nil {
			if os.IsNotExist(err) {
				if util.IsDir(b.Dest) && !strings.HasSuffix(b.Dest, "/") {
					b.Dest += "/"
				}
				overlayObjects = append(overlayObjects, b.Dest)
				wwlog.Debug("Container %s, needs following path: %s", containerName, b.Dest)
			}
		}
	}
	return overlayObjects
}
