//go:build linux
// +build linux

package exec

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/containers/storage/pkg/archive"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/opencontainers/umoci"
	"github.com/opencontainers/umoci/mutate"
	"github.com/opencontainers/umoci/oci/layer"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/oci"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/container"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

var (
	tarReader = errors.New("coudn't create reader for changes")
)

/*
fork off a process with a new PID space
*/
func runContainedCmd(containerName string, args []string) (err error) {
	wwlog.Debug("runContainedCmd: %s %v", containerName, args)
	if len(args) < 1 {
		return fmt.Errorf("runContainedCmd needs at least following args: CMD")
	}
	conf := warewulfconf.Get()
	if overlayDir == "" {
		overlayDir = conf.Paths.WWChrootdir
	}
	if matches, _ := filepath.Glob(path.Join(conf.Paths.WWChrootdir, args[0]) + "-run-*"); len(matches) > 0 {
		return fmt.Errorf("found lock directories for container: %v", matches)
	}
	overlayDir, err = os.MkdirTemp(conf.Paths.WWChrootdir, containerName+"-run-")
	if err != nil {
		wwlog.Warn("couldn't create temp dir for overlay", err)
	}
	defer func() {
		err = errors.Join(os.RemoveAll(overlayDir), err)
	}()
	// find out if changes are recorded on this container
	eng, err := umoci.OpenLayout(warewulfconf.Get().Warewulf.DataStore + "/oci")
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	lst, err := eng.ListReferences(ctx)
	ro := util.InSlice(lst, containerName+oci.CacheContainerSuffix)
	// ro := !util.IsWriteAble(container.RootFsDir(containerName))
	runargs := append([]string{
		"--warewulfconf=" + conf.GetWarewulfConf(),
		"--loglevel=" + fmt.Sprint(wwlog.GetLogLevel()),
		"container", "exec", "__child",
		"--overlaydir=" + overlayDir,
		"--readonly=" + strconv.FormatBool(ro),
		"--containername=" + containerName},
		args...)

	wwlog.Verbose("Running contained command: %s", runargs)
	c := exec.Command("/proc/self/exe", runargs...)
	c.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	err = c.Run()
	if err != nil {
		return err
	}
	if _, err = os.Stat(path.Join(overlayDir, "changes")); err == os.ErrNotExist {
		return nil
	}
	if !ro {
		return nil
	}
	wwlog.Output("Starting to write differences")
	readerBackWrite, err := archive.TarWithOptions(path.Join(overlayDir, "changes"), &archive.TarOptions{
		Compression: archive.Uncompressed})
	if err != nil {
		return errors.Join(err, tarReader)
	}
	/*
			hasher := sha256.New()
		if err != nil {
			return errors.Join(err, errors.New("couldn't tar ball"))
		}
		_, err = io.Copy(hasher, rdHash)
		rdHash.Close()
		if err != nil {
			return errors.Join(err, errors.New("couldn't create hash of changes"))
		}

		file, err := os.Create(path.Join(warewulfconf.Get().Warewulf.DataStore+"/oci/blobs/sha256/", hex.EncodeToString(hasher.Sum(nil))))
		if err != nil {
			return errors.Join(err, errors.New("couldn't open output file"))
		}
		defer file.Close()

		// Copy the data from reader to file, ignore error as we dealt above with it
		rd, _ := archive.TarWithOptions(path.Join(overlayDir, "changes"), &archive.TarOptions{
			Compression: archive.Gzip})
		_, err = io.Copy(file, rd)
		if err != nil {
			return errors.Join(err, errors.New("could't write output"))
		}
		wwlog.Debug("writing back layer: %s -> %s", file.Name(), container.RootFsDir(containerName))
		_, _ = file.Seek(0, 0)
		// we have to uncompress now
		gzR, _ := gzip.NewReader(file)
	*/

	err = layer.UnpackLayer(container.RootFsDir(containerName), readerBackWrite, &layer.UnpackOptions{})
	readerBackWrite.Close()

	if err != nil {
		return errors.Join(err, errors.New("couldn't write back layer"))
	}
	// modify container
	resolvedRef, err := eng.ResolveReference(ctx, containerName+oci.CacheContainerSuffix)
	mut, err := mutate.New(eng, resolvedRef[0])
	comp := mutate.GzipCompressor
	readerMod, err := archive.TarWithOptions(path.Join(overlayDir, "changes"), &archive.TarOptions{
		Compression: archive.Uncompressed})
	if err != nil {
		return errors.Join(err, tarReader)
	}
	defer readerMod.Close()
	t := time.Now()
	hist := v1.History{
		Created:   &t,
		CreatedBy: os.Getenv("LOGNAME"),
		Comment:   strings.Join(runargs, " "),
	}
	desc, err := mut.Add(ctx, v1.MediaTypeImageLayer, readerMod, &hist, comp)
	if err != nil {
		return errors.Join(err, errors.New("couldn't add layer"))
	}
	wwlog.Verbose("update descriptor: %s", desc)
	newPath, err := mut.Commit(ctx)
	if err != nil {
		return errors.Join(err, errors.New("couldn't commit layer"))
	}
	err = eng.UpdateReference(ctx, containerName+oci.CacheContainerSuffix, newPath.Descriptor())
	if err != nil {
		return errors.Join(err, errors.New("couldn't update layer"))
	}
	err = eng.Close()
	if err != nil {
		return errors.Join(err, errors.New("couldn't close engine"))
	}

	/*
		err = os.Chmod(container.RootFsDir(containerName), fs.FileMode(os.O_RDONLY))
		if err != nil {
			return errors.Join(err, fmt.Errorf("couldn't change to ro for: %s", container.RootFsDir(containerName)))
		}
	*/
	return nil
}

func CobraRunE(cmd *cobra.Command, args []string) (err error) {

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
	if nodeName != "" {
		allargs = append(allargs, "--node", nodeName)
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

	err = runContainedCmd(allargs[0], allargs[1:])
	wwlog.Debug("runContainedCmd returned: %v", err)
	if err != nil {
		wwlog.Error("Failed executing container command: %s", err)
		return err
	}

	if util.IsFile(path.Join(container.RootFsDir(allargs[0]), "/etc/warewulf/container_exit.sh")) {
		wwlog.Verbose("Found clean script: /etc/warewulf/container_exit.sh")
		err = runContainedCmd(allargs[0], []string{"/bin/sh", "/etc/warewulf/container_exit.sh"})
		if err != nil {
			wwlog.Error("Failed executing exit script: %s", err)
			return err
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

	wwlog.Output("Rebuilding container...\n")
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

func SetNode(myNode string) {
	nodeName = myNode
}
