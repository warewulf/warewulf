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

	"github.com/spf13/cobra"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/container"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

const exitEval = `$(VALU="$?" ; if [ $VALU == 0 ]; then echo write; else echo discard; fi)`

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	if os.Getpid() != 1 {
		return fmt.Errorf("PID is not 1: %d", os.Getpid())
	}

	containerName := args[0]

	if !container.ValidSource(containerName) {
		return fmt.Errorf("unknown Warewulf container: %s", containerName)
	}
	conf := warewulfconf.Get()
	runDir := container.RunDir(containerName)
	if _, err := os.Stat(runDir); os.IsNotExist(err) {
		return fmt.Errorf("container run directory does not exist: %w", err)
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
		return fmt.Errorf("failed to mount: %w", err)
	}
	ps1Str := fmt.Sprintf("[%s|%s] Warewulf> ", containerName, exitEval)
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
			return fmt.Errorf("could not open node configuration: %s", err)
		}

		allNodes, err := nodeDB.FindAllNodes()
		if err != nil {
			return fmt.Errorf("could not get node list: %s", err)
		}
		filteredNodes := node.FilterNodeListByName(allNodes, []string{nodename})
		if len(filteredNodes) != 1 {
			return fmt.Errorf("no single node idendified with %s", nodename)
		}
		overlays := filteredNodes[0].SystemOverlay
		overlays = append(overlays, filteredNodes[0].RuntimeOverlay...)
		err = overlay.BuildOverlayIndir(filteredNodes[0], allNodes, overlays, path.Join(runDir, "nodeoverlay"))
		if err != nil {
			return fmt.Errorf("could not build overlay: %s", err)
		}
		options := fmt.Sprintf("lowerdir=%s:%s:%s",
			path.Join(runDir, "lower"), containerPath, path.Join(runDir, "nodeoverlay"))
		wwlog.Debug("overlay options: %s", options)
		err = syscall.Mount("overlay", containerPath, "overlay", 0, options)
		if err != nil {
			return fmt.Errorf("Couldn't create overlay for node render overlay: %s", err)
		}
		ps1Str = fmt.Sprintf("[%s|ro|%s] Warewulf> ", containerName, nodename)
	}
	if !container.IsWriteAble(containerName) && nodename == "" {
		wwlog.Verbose("mounting %s ro", containerPath)
		ps1Str = fmt.Sprintf("[%s|ro] Warewulf> ", containerName)
		err = syscall.Mount(containerPath, containerPath, "", syscall.MS_BIND, "")
		if err != nil {
			return fmt.Errorf("failed to prepare bind mount: %w", err)
		}
		err = syscall.Mount(containerPath, containerPath, "", syscall.MS_REMOUNT|syscall.MS_RDONLY|syscall.MS_BIND, "")
		if err != nil {
			return fmt.Errorf("failed to remount ro: %w", err)
		}
	}

	for _, mntPnt := range mountPts {
		if mntPnt.Copy() {
			continue
		}
		wwlog.Debug("bind mounting: %s -> %s", mntPnt.Source, path.Join(containerPath, mntPnt.Dest))
		err = syscall.Mount(mntPnt.Source, path.Join(containerPath, mntPnt.Dest), "", syscall.MS_BIND, "")
		if err != nil {
			wwlog.Warn("Couldn't mount %s to %s: %s", mntPnt.Source, mntPnt.Dest, err)
		} else if mntPnt.ReadOnly() {
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
		return fmt.Errorf("failed to chroot: %w", err)
	}

	err = os.Chdir("/")
	if err != nil {
		return fmt.Errorf("failed to chdir: %w", err)
	}

	if err := syscall.Mount("devtmpfs", "/dev", "devtmpfs", 0, ""); err != nil {
		return fmt.Errorf("failed to mount /dev: %w", err)
	}
	if err := syscall.Mount("sysfs", "/sys", "sysfs", 0, ""); err != nil {
		return fmt.Errorf("failed to mount /sys: %w", err)
	}
	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		return fmt.Errorf("failed to mount /proc: %w", err)
	}
	if err := syscall.Mount("tmpfs", "/run", "tmpfs", 0, ""); err != nil {
		return fmt.Errorf("failed to mount /run: %w", err)
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
		if b.Copy() {
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
