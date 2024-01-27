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
	mountPts := conf.MountsContainer
	mountPts = append(container.InitMountPnts(binds), mountPts...)
	// check for valid mount points
	lowerObjects := checkMountPoints(containerName, mountPts)
	overlayDir := conf.Paths.WWChrootdir + "/overlays"
	// need to create a overlay, where the lower layer contains
	// the missing mount points
	wwlog.Verbose("for ephermal mount use tempdir %s", overlayDir)
	// ignore errors as we are doomed if a tmp dir couldn't be written
	_ = os.MkdirAll(path.Join(overlayDir, "work"), os.ModePerm)
	_ = os.MkdirAll(path.Join(overlayDir, "lower"), os.ModePerm)
	_ = os.MkdirAll(path.Join(overlayDir, "nodeoverlay"), os.ModePerm)
	for _, obj := range lowerObjects {
		newFile := ""
		if !strings.HasSuffix(obj, "/") {
			newFile = filepath.Base(obj)
			obj = filepath.Dir(obj)
		}
		err = os.MkdirAll(filepath.Join(overlayDir, "lower", obj), os.ModePerm)
		if err != nil {
			wwlog.Warn("couldn't create directory for mounts: %s", err)
		}
		if newFile != "" {
			desc, err := os.Create(filepath.Join(overlayDir, "lower", obj, newFile))
			if err != nil {
				wwlog.Warn("couldn't create directory for mounts: %s", err)
			}
			defer desc.Close()
		}
	}
	containerPath := container.RootFsDir(containerName)
	err = syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	if err != nil {
		return errors.Wrap(err, "failed to mount")
	}
	ps1Str := fmt.Sprintf("[%s] Warewulf> ", containerName)
	if len(lowerObjects) != 0 && nodename == "" {
		options := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s",
			path.Join(overlayDir, "lower"), containerPath, path.Join(overlayDir, "work"))
		wwlog.Debug("overlay options: %s", options)
		err = syscall.Mount("overlay", containerPath, "overlay", 0, options)
		if err != nil {
			wwlog.Warn(fmt.Sprintf("Couldn't create overlay for ephermal mount points: %s", err))
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
		overlays := nodes[0].SystemOverlay.GetSlice()
		overlays = append(overlays, nodes[0].RuntimeOverlay.GetSlice()...)
		err = overlay.BuildOverlayIndir(nodes[0], overlays, path.Join(overlayDir, "nodeoverlay"))
		if err != nil {
			wwlog.Error("Could not build overlay: %s", err)
			os.Exit(1)
		}
		options := fmt.Sprintf("lowerdir=%s:%s:%s",
			path.Join(overlayDir, "lower"), containerPath, path.Join(overlayDir, "nodeoverlay"))
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

	err = syscall.Mount("/dev", path.Join(containerPath, "/dev"), "", syscall.MS_BIND, "")
	if err != nil {
		return errors.Wrap(err, "failed to mount /dev")
	}

	for _, mntPnt := range mountPts {
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

	err = syscall.Mount("/proc", "/proc", "proc", 0, "")
	if err != nil {
		return errors.Wrap(err, "failed to mount proc")
	}

	os.Setenv("PS1", ps1Str)
	os.Setenv("PATH", "/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin")
	os.Setenv("HISTFILE", "/dev/null")

	_ = syscall.Exec(args[1], args[1:], os.Environ())
	/*
		Exec replaces the actual program, so nothing to do here afterwards
	*/
	return nil
}

/*
Check if the bind mount points exists in the given container. Returns
the invalid mount points. Directories always have '/' as suffix
*/
func checkMountPoints(containerName string, binds []*warewulfconf.MountEntry) (overlayObjects []string) {
	overlayObjects = []string{}
	for _, b := range binds {
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
