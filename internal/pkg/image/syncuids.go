package image

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
// image identified by imageName by installing the equivalent
// files from the host and appending names only in the
// image. Files in the image are updated to match the new
// numeric id assignments.
//
// If write is false, the image is not actually updated, but
// relevant log entries and output are generated.
//
// Any I/O errors are returned unmodified.
//
// A conflict arises if the image has an entry with the same id as
// an entry in the host and the host does not have an entry with the
// same name. In this case, an error is returned.
func SyncUids(imageName string, write bool) error {
	wwlog.Debug("SyncUids(imageName=%v, write=%v)", imageName, write)
	imagePath := RootFsDir(imageName)
	imagePasswdPath := path.Join(imagePath, passwdPath)
	imageGroupPath := path.Join(imagePath, groupPath)

	passwdSync := make(syncDB)
	if err := passwdSync.readFromHost(passwdPath); err != nil {
		return err
	}
	if err := passwdSync.readFromimage(imagePasswdPath); err != nil {
		return err
	}
	if err := passwdSync.checkConflicts(); err != nil {
		return err
	}

	groupSync := make(syncDB)
	if err := groupSync.readFromHost(groupPath); err != nil {
		return err
	}
	if err := groupSync.readFromimage(imageGroupPath); err != nil {
		return err
	}
	if err := groupSync.checkConflicts(); err != nil {
		return err
	}

	if err := passwdSync.findUserFiles(imagePath); err != nil {
		return err
	}
	if err := groupSync.findGroupFiles(imagePath); err != nil {
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
		if err := passwdSync.update(imagePasswdPath, passwdPath); err != nil {
			return err
		}
		if err := groupSync.update(imageGroupPath, groupPath); err != nil {
			return err
		}
		wwlog.Info("uid/gid synced for image %s", imageName)
	} else {
		if passwdSync.needsSync() || groupSync.needsSync() {
			wwlog.Info("uid/gid not synced: run `wwctl image syncuser --write %s`", imageName)
		} else {
			wwlog.Info("uid/gid already synced")
		}
	}

	return nil
}

// A syncDB maps user or group names to syncInfo instances, which
// correlate IDs between host and image and track affected
// files. This can be used for either /etc/passwd or /etc/group IDs.
type syncDB map[string]syncInfo

// checkConflicts inspects a syncDB map for irreconcilable
// conflicts. A conflict arises if the image has an entry with the
// same id as an entry in the host and the host does not have an entry
// with the same name.
func (db syncDB) checkConflicts() error {
	for nameInimage, imageIds := range db {
		if !imageIds.inimage() || imageIds.inHost() {
			continue
		}

		for nameInHost, hostIds := range db {
			if hostIds.HostID == imageIds.imageID {
				wwlog.Warn("syncuser cannot determine what name should be assigned to id number %v: on the local host it is assigned to '%s'; but in the image it is assigned to '%s'. Since '%s' does not exist on the local host, it cannot be reassigned to a different id number. To resolve the conflict, add '%s' to the local host and re-run syncuser.", imageIds.imageID, nameInHost, nameInimage, nameInimage, nameInimage)
				return errors.New(fmt.Sprintf("id(%v) collision: host(%s) image(%s)", imageIds.imageID, nameInHost, nameInimage))
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
	var onlyimage, onlyHost, match, differ []string
	for name, ids := range db {
		wwlog.Debug("%s: %s: host(%v) image(%v) files: %v", prefix, name, ids.HostID, ids.imageID, ids.imageFiles)
		if ids.onlyimage() {
			onlyimage = append(onlyimage, name)
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

	wwlog.Verbose("%s: image only: %v", prefix, onlyimage)
	wwlog.Verbose("%s: host only: %v", prefix, onlyHost)
	wwlog.Verbose("%s: match: %v", prefix, match)
	wwlog.Verbose("%s: differ: %v", prefix, differ)
}

// read reads fileName (an /etc/passwd or /etc/group file) into a
// syncDB map. Since the name and numerical id for both files are
// stored at fields 0 and 2, the same function can read both files.
//
// fromimage identifies whether the given fileName includes
// entries from the host or the image.
func (db syncDB) read(fileName string, fromimage bool) error {
	wwlog.Debug("read %s (image: %t)", fileName, fromimage)
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
			// ignore ldap/nis/sssd line
			if strings.HasPrefix(fields[0], "+") || strings.HasPrefix(fields[0], "-") {
				wwlog.Verbose("Ignoring line %s (unhandled compat-style NIS reference)", line)
				continue
			}
			id, err := strconv.Atoi(fields[2])
			if err != nil {
				wwlog.Warn("Ignoring line %s (parse error)", line)
				continue
			}
			entry, ok := db[name]
			if !ok {
				entry = syncInfo{HostID: -1, imageID: -1}
			}

			if fromimage {
				if entry.imageID == -1 {
					entry.imageID = id
				} else if entry.imageID != id {
					wwlog.Warn("Ignoring image id(%v) for %s from %s after existing value %v", id, name, fileName, entry.imageID)
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

// readFromimage reads fileName into a syncDB map.
//
// Equivalent to read(fileName, true)
func (db syncDB) readFromimage(fileName string) error { return db.read(fileName, true) }

// findFiles finds files under imagePath that are owned by each
// tracked image ID.
//
// If byGid is true, files with a matching gid are
// recorded. Otherwise, files with a matching uid are recorded.
func (db syncDB) findFiles(imagePath string, byGid bool) error {
	wwlog.Debug("findFiles(imagePath=%v, byGid=%v)", imagePath, byGid)
	syncIDs := make(map[int]string)
	for name, info := range db {
		if info.inHost() && info.inimage() && !info.match() {
			wwlog.Debug("syncID[%v] = %v", info.imageID, name)
			syncIDs[info.imageID] = name
		}
	}

	return filepath.Walk(imagePath, func(filePath string, fileInfo fs.FileInfo, err error) error {
		if stat, ok := fileInfo.Sys().(*syscall.Stat_t); ok {
			var id int
			if byGid {
				id = int(stat.Gid)
			} else {
				id = int(stat.Uid)
			}
			if name, ok := syncIDs[id]; ok {
				info := db[name]
				wwlog.Debug("findFiles: %s: (%v -> %v, gid: %v)", filePath, info.imageID, info.HostID, byGid)
				info.imageFiles = append(info.imageFiles, filePath)
				db[name] = info
			} else {
				wwlog.Debug("findFiles: %s", filePath)
			}
		}
		return nil
	})
}

// findUserFiles is equivalent to findFiles(imagePath, false)
func (db syncDB) findUserFiles(imagePath string) error { return db.findFiles(imagePath, false) }

// findGroupFiles is equivalent to findFiles(imagePath, true)
func (db syncDB) findGroupFiles(imagePath string) error { return db.findFiles(imagePath, true) }

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

// getOnlyimageLines returns the lines from fileName (either an
// /etc/passwd or /etc/group file) for names that occur only in the
// image.
//
// These lines are added to the respective file from the host when
// updating the image.
func (db syncDB) getOnlyimageLines(fileName string) ([]string, error) {
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
				if ids.onlyimage() {
					lines = append(lines, line)
				}
				break
			}
		}
	}
	wwlog.Debug("file: %s, entries only in image: %v", fileName, lines)
	return lines, nil
}

// update replaces imagePath with hostPath and adds lines from
// getOnlyimageLines to the end of imagePath. This is used to
// replace /etc/passwd (or /etc/group) in the image with the
// equivalent file from the host while retaining names unique to the
// image.
func (db syncDB) update(imagePath string, hostPath string) error {
	wwlog.Debug("update %s from %s)", imagePath, hostPath)
	if lines, err := db.getOnlyimageLines(imagePath); err != nil {
		return err
	} else {
		if err := os.Remove(imagePath); err != nil {
			return err
		}
		if err := util.CopyFile(hostPath, imagePath); err != nil {
			return err
		}
		if err := util.AppendLines(imagePath, lines); err != nil {
			return err
		}
		return nil
	}
}

// needsSync returns true if the syncDB map indicates that ids between
// the image and host are out-of-sync.
func (db syncDB) needsSync() bool {
	for name, ids := range db {
		if ids.onlyHost() {
			wwlog.Debug("sync required: %s only in host", name)
			return true
		}
		if ids.differ() {
			wwlog.Debug("sync required: %s is %v in host and %v in image", name, ids.HostID, ids.imageID)
			return true
		}
	}
	return false
}

// sycncInfo correlates the numerical id of a name on the host
// (HostID) and the image (imageID), along with the files in
// the image that are owned by that name. This allows affected
// files to be updated when the HostID is applied to the image.
type syncInfo struct {
	HostID     int      `access:"r,w"`
	imageID    int      `access:"r,w"`
	imageFiles []string `access:"r,w"`
}

// inHost returns true if info has a record of an id for this name in
// the host.
func (info *syncInfo) inHost() bool {
	return info.HostID != -1
}

// inimage returns true if info has a record of an id for this
// name in the image.
func (info *syncInfo) inimage() bool {
	return info.imageID != -1
}

// onlyHost returns true if info has a record of an id for this name
// in the host and not in the image.
func (info *syncInfo) onlyHost() bool {
	return info.inHost() && !info.inimage()
}

// onlyHost returns true if info has a record of an id for this name
// in the image and not in the host.
func (info *syncInfo) onlyimage() bool {
	return info.inimage() && !info.inHost()
}

// match returns true if info represents a name that exists with the
// same numerical id in both the host and the image.
func (info *syncInfo) match() bool {
	return info.inimage() && info.inHost() && info.HostID == info.imageID
}

// differ returns true if info represents a name that exists in both
// the host and the image but with different numerical ids.
func (info *syncInfo) differ() bool {
	return info.inimage() && info.inHost() && info.HostID != info.imageID
}

// chownFiles updates the files recorded in info.imageFiles to use
// the numerical IDs from the host.
//
// If byGid is true, the file's gid is updated; otherwise, the file's
// uid is updated.
func (info *syncInfo) chownFiles(byGid bool) error {
	for _, file := range info.imageFiles {
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
