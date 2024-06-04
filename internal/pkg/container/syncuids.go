package container

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

const passwdPath = "/etc/passwd"
const groupPath = "/etc/group"

// SyncUids updates the /etc/passwd and /etc/group files in the
// container identified by containerName by installing the equivalent
// files from the host and appending names only in the
// container. Files in the container are updated to match the new
// numeric id assignments.
//
// If write is false, the container is not actually updated, but
// relevant log entries and output are generated.
//
// Any I/O errors are returned unmodified.
//
// A conflict arises if the container has an entry with the same id as
// an entry in the host and the host does not have an entry with the
// same name. In this case, an error is returned.
func SyncUids(containerName string, write bool) error {
	wwlog.Debug("SyncUids(containerName=%v, write=%v)", containerName, write)
	containerPath := RootFsDir(containerName)
	containerPasswdPath := path.Join(containerPath, passwdPath)
	containerGroupPath := path.Join(containerPath, groupPath)

	passwdSync := make(syncDB)
	if err := passwdSync.readFromHost(passwdPath); err != nil {
		return err
	}
	if err := passwdSync.readFromContainer(containerPasswdPath); err != nil {
		return err
	}
	if err := passwdSync.checkConflicts(); err != nil {
		return err
	}

	groupSync := make(syncDB)
	if err := groupSync.readFromHost(groupPath); err != nil {
		return err
	}
	if err := groupSync.readFromContainer(containerGroupPath); err != nil {
		return err
	}
	if err := groupSync.checkConflicts(); err != nil {
		return err
	}

	if err := passwdSync.findUserFiles(containerPath); err != nil {
		return err
	}
	if err := groupSync.findGroupFiles(containerPath); err != nil {
		return err
	}

	passwdSync.log("passwd")
	groupSync.log("group")

	if write {
		if err := passwdSync.chownUserFiles(); err != nil {
			return err
		}
		if err := groupSync.chownGroupFiles(); err != nil {
			return err
		}
		if err := passwdSync.update(containerPasswdPath, passwdPath); err != nil {
			return err
		}
		if err := groupSync.update(containerGroupPath, groupPath); err != nil {
			return err
		}
		wwlog.Info("uid/gid synced for container %s", containerName)
	} else {
		if passwdSync.needsSync() || groupSync.needsSync() {
			wwlog.Info("uid/gid not synced: run `wwctl container syncuser --write %s`", containerName)
		} else {
			wwlog.Info("uid/gid already synced")
		}
	}

	return nil
}

// A syncDB maps user or group names to syncInfo instances, which
// correlate IDs between host and container and track affected
// files. This can be used for either /etc/passwd or /etc/group IDs.
type syncDB map[string]syncInfo

// checkConflicts inspects a syncDB map for irreconcilable
// conflicts. A conflict arises if the container has an entry with the
// same id as an entry in the host and the host does not have an entry
// with the same name.
func (db syncDB) checkConflicts() error {
	for nameInContainer, containerIds := range db {
		if !containerIds.inContainer() || containerIds.inHost() {
			continue
		}

		for nameInHost, hostIds := range db {
			if hostIds.HostID == containerIds.ContainerID {
				errorMsg := fmt.Sprintf("id(%v) collision: host(%s) container(%s)", containerIds.ContainerID, nameInHost, nameInContainer)
				wwlog.Warn(errorMsg)
				wwlog.Warn("add %s to host to resolve conflict", nameInContainer)
				return errors.New(errorMsg)
			}
		}
	}
	return nil
}

// log generates debug and verbose logs inspecting a syncDB map.
//
// The provided prefix is prepended to log entries to provide context
// for the given syncDB map. (e.g., to differentiate between a user or
// group map)
func (db syncDB) log(prefix string) {
	var onlyContainer, onlyHost, match, differ []string
	for name, ids := range db {
		wwlog.Debug("%s: %s: host(%v) container(%v) files: %v", prefix, name, ids.HostID, ids.ContainerID, ids.ContainerFiles)
		if ids.onlyContainer() {
			onlyContainer = append(onlyContainer, name)
		}
		if ids.onlyHost() {
			onlyHost = append(onlyHost, name)
		}
		if ids.match() {
			match = append(match, name)
		}
		if ids.differ() {
			differ = append(differ, name)
		}
	}

	wwlog.Verbose("%s: container only: %v", prefix, onlyContainer)
	wwlog.Verbose("%s: host only: %v", prefix, onlyHost)
	wwlog.Verbose("%s: match: %v", prefix, match)
	wwlog.Verbose("%s: differ: %v", prefix, differ)
}

// read reads fileName (an /etc/passwd or /etc/group file) into a
// syncDB map. Since the name and numerical id for both files are
// stored at fields 0 and 2, the same function can read both files.
//
// fromContainer identifies whether the given fileName includes
// entries from the host or the container.
func (db syncDB) read(fileName string, fromContainer bool) error {
	wwlog.Debug("read %s (container: %t)", fileName, fromContainer)
	if file, err := os.Open(fileName); err != nil {
		return err
	} else {
		defer file.Close()
		fileScanner := bufio.NewScanner(file)
		for fileScanner.Scan() {
			line := fileScanner.Text()
			fields := strings.Split(line, ":")
			if len(fields) != 7 && len(fields) != 4 {
				wwlog.Debug("malformed line in passwd/group: %s", line)
				continue
			}
			name := fields[0]
			if name == "" {
				continue
			}

			id, err := strconv.Atoi(fields[2])
			if err != nil {
				wwlog.Warn("Ignoring line %s (parse error)", line)
				continue
			}

			entry, ok := db[name]
			if !ok {
				entry = syncInfo{HostID: -1, ContainerID: -1}
			}

			if fromContainer {
				if entry.ContainerID == -1 {
					entry.ContainerID = id
				} else if entry.ContainerID != id {
					wwlog.Warn("Ignoring container id(%v) for %s from %s after existing value %v", id, name, fileName, entry.ContainerID)
					continue
				}
			} else {
				if entry.HostID == -1 {
					entry.HostID = id
				} else if entry.HostID != id {
					wwlog.Warn("Ignoring host id(%v) for %s from %s after existing value %v", id, name, fileName, entry.HostID)
					continue
				}
			}

			db[name] = entry
		}
		return nil
	}
}

// readFromHost reads fileName into a syncDB map.
//
// Equivalent to read(fileName, false)
func (db syncDB) readFromHost(fileName string) error { return db.read(fileName, false) }

// readFromContainer reads fileName into a syncDB map.
//
// Equivalent to read(fileName, true)
func (db syncDB) readFromContainer(fileName string) error { return db.read(fileName, true) }

// findFiles finds files under containerPath that are owned by each
// tracked container ID.
//
// If byGid is true, files with a matching gid are
// recorded. Otherwise, files with a matching uid are recorded.
func (db syncDB) findFiles(containerPath string, byGid bool) error {
	wwlog.Debug("findFiles(containerPath=%v, byGid=%v)", containerPath, byGid)
	syncIDs := make(map[int]string)
	for name, info := range db {
		if info.inHost() && info.inContainer() && !info.match() {
			wwlog.Debug("syncID[%v] = %v", info.ContainerID, name)
			syncIDs[info.ContainerID] = name
		}
	}

	return filepath.Walk(containerPath, func(filePath string, fileInfo fs.FileInfo, err error) error {
		if stat, ok := fileInfo.Sys().(*syscall.Stat_t); ok {
			var id int
			if byGid {
				id = int(stat.Gid)
			} else {
				id = int(stat.Uid)
			}
			if name, ok := syncIDs[id]; ok {
				info := db[name]
				wwlog.Debug("findFiles: %s: (%v -> %v, gid: %v)", filePath, info.ContainerID, info.HostID, byGid)
				info.ContainerFiles = append(info.ContainerFiles, filePath)
				db[name] = info
			} else {
				wwlog.Debug("findFiles: %s", filePath)
			}
		}
		return nil
	})
}

// findUserFiles is equivalent to findFiles(containerPath, false)
func (db syncDB) findUserFiles(containerPath string) error { return db.findFiles(containerPath, false) }

// findGroupFiles is equivalent to findFiles(containerPath, true)
func (db syncDB) findGroupFiles(containerPath string) error { return db.findFiles(containerPath, true) }

// chownFiles updates found files (see findFiles) to reflect ownership
// using host ids.
//
// If byGid is true, file gids are updated. Otherwise, file uids are
// updated.
func (db syncDB) chownFiles(byGid bool) error {
	for _, ids := range db {
		if err := ids.chownFiles(byGid); err != nil {
			return err
		}
	}
	return nil
}

// chownUserFiles is equivalent to chownFiles(false)
func (db syncDB) chownUserFiles() error { return db.chownFiles(false) }

// chownUserFiles is equivalent to chownFiles(true)
func (db syncDB) chownGroupFiles() error { return db.chownFiles(true) }

// getOnlyContainerLines returns the lines from fileName (either an
// /etc/passwd or /etc/group file) for names that occur only in the
// container.
//
// These lines are added to the respective file from the host when
// updating the container.
func (db syncDB) getOnlyContainerLines(fileName string) ([]string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fileScanner := bufio.NewScanner(file)

	var lines []string
	for fileScanner.Scan() {
		line := fileScanner.Text()
		fields := strings.Split(line, ":")
		for name, ids := range db {
			if fields[0] == name {
				if ids.onlyContainer() {
					lines = append(lines, line)
				}
				break
			}
		}
	}
	wwlog.Debug("file: %s, entries only in container: %v", fileName, lines)
	return lines, nil
}

// update replaces containerPath with hostPath and adds lines from
// getOnlyContainerLines to the end of containerPath. This is used to
// replace /etc/passwd (or /etc/group) in the container with the
// equivalent file from the host while retaining names unique to the
// container.
func (db syncDB) update(containerPath string, hostPath string) error {
	wwlog.Debug("update %s from %s)", containerPath, hostPath)
	if lines, err := db.getOnlyContainerLines(containerPath); err != nil {
		return err
	} else {
		if err := os.Remove(containerPath); err != nil {
			return err
		}
		if err := util.CopyFile(hostPath, containerPath); err != nil {
			return err
		}
		if err := util.AppendLines(containerPath, lines); err != nil {
			return err
		}
		return nil
	}
}

// needsSync returns true if the syncDB map indicates that ids between
// the container and host are out-of-sync.
func (db syncDB) needsSync() bool {
	for name, ids := range db {
		if ids.onlyHost() {
			wwlog.Debug("sync required: %s only in host", name)
			return true
		}
		if ids.differ() {
			wwlog.Debug("sync required: %s is %v in host and %v in container", name, ids.HostID, ids.ContainerID)
			return true
		}
	}
	return false
}

// sycncInfo correlates the numerical id of a name on the host
// (HostID) and the container (ContainerID), along with the files in
// the container that are owned by that name. This allows affected
// files to be updated when the HostID is applied to the container.
type syncInfo struct {
	HostID         int      `access:"r,w"`
	ContainerID    int      `access:"r,w"`
	ContainerFiles []string `access:"r,w"`
}

// inHost returns true if info has a record of an id for this name in
// the host.
func (info *syncInfo) inHost() bool {
	return info.HostID != -1
}

// inContainer returns true if info has a record of an id for this
// name in the container.
func (info *syncInfo) inContainer() bool {
	return info.ContainerID != -1
}

// onlyHost returns true if info has a record of an id for this name
// in the host and not in the container.
func (info *syncInfo) onlyHost() bool {
	return info.inHost() && !info.inContainer()
}

// onlyHost returns true if info has a record of an id for this name
// in the container and not in the host.
func (info *syncInfo) onlyContainer() bool {
	return info.inContainer() && !info.inHost()
}

// match returns true if info represents a name that exists with the
// same numerical id in both the host and the container.
func (info *syncInfo) match() bool {
	return info.inContainer() && info.inHost() && info.HostID == info.ContainerID
}

// differ returns true if info represents a name that exists in both
// the host and the container but with different numerical ids.
func (info *syncInfo) differ() bool {
	return info.inContainer() && info.inHost() && info.HostID != info.ContainerID
}

// chownFiles updates the files recorded in info.ContainerFiles to use
// the numerical IDs from the host.
//
// If byGid is true, the file's gid is updated; otherwise, the file's
// uid is updated.
func (info *syncInfo) chownFiles(byGid bool) error {
	for _, file := range info.ContainerFiles {
		if fileInfo, err := os.Stat(file); err != nil {
			return err
		} else {
			if fileInfo.IsDir() || fileInfo.Mode().IsRegular() {
				var newUid, newGid int
				if byGid {
					newUid = -1
					newGid = info.HostID
				} else {
					newUid = info.HostID
					newGid = -1
				}
				if err := os.Chown(file, newUid, newGid); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
