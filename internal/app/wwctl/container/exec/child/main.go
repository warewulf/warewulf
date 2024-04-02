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

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	if os.Getpid() != 1 {
		wwlog.Error("PID is not 1: %d", os.Getpid())
		os.Exit(1)
	}

	if !container.ValidSource(containerName) {
		wwlog.Error("Unknown Warewulf container: %s", containerName)
		os.Exit(1)
	}
	conf := warewulfconf.Get()
	if overlayDir == "" {
		overlayDir = path.Join(conf.Paths.WWChrootdir, "overlays")
	}
	mountPts := conf.MountsContainer
	mountPts = append(container.InitMountPnts(binds), mountPts...)
	// check for valid mount points
	lowerObjects := checkMountPoints(containerName, mountPts)
	// need to create a overlay, where the lower layer contains
	// the missing mount points
	wwlog.Verbose("for ephermal mount use tempdir %s", overlayDir)
	// ignore errors as we are doomed if a tmp dir couldn't be written
	_ = os.MkdirAll(path.Join(overlayDir, "work"), os.ModePerm)
	_ = os.MkdirAll(path.Join(overlayDir, "lower"), os.ModePerm)
	_ = os.MkdirAll(path.Join(overlayDir, "nodeoverlay"), os.ModePerm)
	// handle all lower object, have some extra logic if the object is a file
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
	// running in a private PID space, so also make / private, so that nothing gets out from here
	err = syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	if err != nil {
		return errors.Wrap(err, "failed to mount")
	}
	ps1Str := fmt.Sprintf("[%s|%s] Warewulf> ", exitEval, containerName)
	msgStr := `Image is rebuilt, depending on the exit status of the last called program.
Type "true" or "false" to enforce or abort image rebuilt.`
	if len(lowerObjects) != 0 && nodename == "" && !recordChanges {
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
			return err
		}
		node, err := nodeDB.FindById(nodename)
		if err != nil {
			return err
		}
		overlays := node.SystemOverlay.GetSlice()
		overlays = append(overlays, node.RuntimeOverlay.GetSlice()...)
		err = overlay.BuildOverlayIndir(node, overlays, path.Join(overlayDir, "nodeoverlay"))
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
	} else if recordChanges && nodename == "" {
		// if !util.IsWriteAble(containerPath) && nodename == "" {
		_ = os.MkdirAll(path.Join(overlayDir, "changes"), os.ModePerm)
		ps1Str = fmt.Sprintf("[%s|%s] Warewulf> ", exitEval, containerName)
		options := fmt.Sprintf("nfs_export=off,lowerdir=%s,upperdir=%s,workdir=%s",
			path.Join(overlayDir, "lower")+":"+containerPath,
			path.Join(overlayDir, "changes"), path.Join(overlayDir, "work"))
		wwlog.Debug("overlay options: %s", options)
		err = syscall.Mount("overlay", containerPath, "overlay", 0, options)
		if err != nil {
			return errors.Wrap(err, "failed to prepare mount")
		}
		msgStr = `Changes are written back to container and image is rebuilt
depending on exit status of last called program.
Type "true" or "false" to enforce or abort image rebuilt.`
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
	fmt.Println(msgStr)
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

	err = syscall.Exec(args[0], args, os.Environ())
	wwlog.Debug("Exec ended with: %v", err)
	return err
	/*
		Exec replaces the actual program, so nothing to do here afterwards
	*/
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
