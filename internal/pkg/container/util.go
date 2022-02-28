package container

import (
	"bufio"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/pkg/errors"

	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

func ValidName(name string) bool {
	if !util.ValidString(name, "^[\\w\\-\\.\\:]+$") {
		wwlog.Printf(wwlog.WARN, "VNFS name has illegal characters: %s\n", name)
		return false
	}
	return true
}

func ListSources() ([]string, error) {
	var ret []string

	err := os.MkdirAll(SourceParentDir(), 0755)
	if err != nil {
		return ret, errors.New("Could not create VNFS source parent directory: " + SourceParentDir())
	}
	wwlog.Printf(wwlog.DEBUG, "Searching for VNFS Rootfs directories: %s\n", SourceParentDir())

	sources, err := ioutil.ReadDir(SourceParentDir())
	if err != nil {
		return ret, err
	}

	for _, source := range sources {
		wwlog.Printf(wwlog.VERBOSE, "Found VNFS source: %s\n", source.Name())

		if !ValidName(source.Name()) {
			continue
		}

		if !ValidSource(source.Name()) {
			continue
		}

		ret = append(ret, source.Name())
	}

	return ret, nil
}

func ValidSource(name string) bool {
	fullPath := RootFsDir(name)

	if !ValidName(name) {
		return false
	}

	if !util.IsDir(fullPath) {
		wwlog.Printf(wwlog.VERBOSE, "Location is not a VNFS source directory: %s\n", name)
		return false
	}

	return true
}

func DeleteSource(name string) error {
	fullPath := SourceDir(name)

	wwlog.Printf(wwlog.VERBOSE, "Removing path: %s\n", fullPath)
	return os.RemoveAll(fullPath)
}

type completeUserInfo struct {
	Name        string
	UidHost     int      `access:"r,w"`
	GidHost     int      `access:"r,w"`
	UidCont     int      `access:"r,w"`
	GidCont     int      `access:"r,w"`
	FileListUid []string `access:"r,w"`
	FileListGid []string `access:"r,w"`
}

type simpleUserInfo struct {
	name string
	uid  int
	gid  int
}

/*
sync the uids,gids from the host to the container
*/
func SyncUids(name string) error {
	var userDb []completeUserInfo
	passwdName := "/etc/passwd"
	groupName := "/etc/group"
	fullPath := RootFsDir(name)
	hostName, err := createPasswdMap(passwdName)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open "+passwdName)
		return err
	}
	// populate db with the user of the
	for _, user := range hostName {
		userDb = append(userDb, completeUserInfo{Name: user.name,
			UidHost: user.uid, GidHost: user.gid, UidCont: -1, GidCont: -1})
	}

	contName, err := createPasswdMap(path.Join(fullPath, passwdName))
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open "+path.Join(fullPath, passwdName))
		return err
	}
	for _, userCont := range contName {
		foundUser := false
		for idxHost, userHost := range userDb {
			if userCont.name == userHost.Name {
				foundUser = true
				(&userDb[idxHost]).UidCont = userCont.uid
				(&userDb[idxHost]).GidCont = userCont.gid
			}
		}
		if !foundUser {
			userDb = append(userDb, completeUserInfo{Name: userCont.name,
				UidHost: -1, GidHost: -1, UidCont: userCont.uid, GidCont: userCont.gid})
			wwlog.Printf(wwlog.WARN, "user: %s:%v:%v not present on host", userCont.name, userCont.uid, userCont.gid)
		}

	}
	// find out which user/goup are only in the container
	var userOnlyCont []string
	for _, user := range userDb {
		if user.UidHost == -1 {
			userOnlyCont = append(userOnlyCont, user.Name)
			for _, userCheck := range userDb {
				if userCheck.UidHost == user.UidCont {
					wwlog.Printf(wwlog.WARN, fmt.Sprintf("uid(%v) collision for host: %s and container: %s\n",
						user.UidCont, user.Name, userCheck.Name))
					return errors.New(fmt.Sprintf("user %s only present in container has same uid(%v) as user %s on host,\n"+
						"add this user to /etc/passwd on host", user.Name, user.UidCont, userCheck.Name))
				}
			}
		}
		if user.GidHost == -1 {
			for _, userCheck := range userDb {
				if userCheck.GidHost == user.GidCont {
					wwlog.Printf(wwlog.WARN, fmt.Sprintf("gid(%v) collision for host: %s and container: %s\n",
						user.GidCont, user.Name, userCheck.Name))
					return errors.New(fmt.Sprintf("user %s only present in container has same gid(%v) as user %s on host,\n"+
						" add this group to /etc/group on host", user.Name, user.GidCont, userCheck.Name))
				}
			}
		}

	}
	// create list of files which need changed ownerships in order to change them later what
	// avoid uid/gid collisions
	for idx, user := range userDb {
		if (user.UidHost != user.UidCont && user.UidCont != -1) || (user.GidHost != user.GidCont && user.GidCont != -1) {
			wwlog.Printf(wwlog.VERBOSE, fmt.Sprintf("host %s:%v:%v <-> container %s:%v:%v\n",
				user.Name, user.UidHost, user.GidHost, user.Name, user.UidCont, user.GidCont))
			err = filepath.Walk(fullPath, func(filePath string, info fs.FileInfo, err error) error {
				// root is always good, if we failt to get UID/GID of a file
				var uid, gid int
				if stat, ok := info.Sys().(*syscall.Stat_t); ok {
					uid = int(stat.Uid)
					gid = int(stat.Gid)
				}
				if uid == user.UidCont {
					(&userDb[idx]).FileListUid = append((&userDb[idx]).FileListUid, filePath)
				}
				if gid == user.GidCont {
					(&userDb[idx]).FileListGid = append((&userDb[idx]).FileListGid, filePath)
				}
				return nil
			})
		}
	}
	// change uids and gid of file
	for _, user := range userDb {
		if len(user.FileListUid) != 0 {
			fmt.Printf("uidList(%s): %v\n", user.Name, user.FileListUid)
			for _, file := range user.FileListUid {
				fsInfo, err := os.Stat(file)
				if err != nil {
					return err
				}
				var gid int
				if stat, ok := fsInfo.Sys().(*syscall.Stat_t); ok {
					gid = int(stat.Gid)
				}
				wwlog.Printf(wwlog.DEBUG, "%s chown(%v,%v)\n", file, user.UidHost, gid)
				err = os.Chown(file, user.UidHost, gid)
				if err != nil {
					return err
				}
			}
		}
		if len(user.FileListGid) != 0 {
			fmt.Printf("gidList(%s): %v\n", user.Name, user.FileListGid)
			for _, file := range user.FileListGid {
				fsInfo, err := os.Stat(file)
				if err != nil {
					return err
				}
				var uid int
				if stat, ok := fsInfo.Sys().(*syscall.Stat_t); ok {
					uid = int(stat.Uid)
				}
				wwlog.Printf(wwlog.DEBUG, "%s chown(%v,%v)\n", file, user.UidHost, uid)
				err = os.Chown(file, uid, user.GidHost)
				if err != nil {
					return err
				}
			}

		}

	}
	// get the entries for the passwd/group file before copy over
	passwdEntries, err := getEntires(path.Join(fullPath, passwdName), userOnlyCont)
	if err != nil {
		return err
	}
	// implicitly assuming that users/groups which only exists on the host have the same name
	groupEntries, err := getEntires(path.Join(fullPath, groupName), userOnlyCont)
	if err != nil {
		return err
	}
	if err = util.CopyFile(passwdName, path.Join(fullPath, passwdName)); err != nil {
		return err
	}
	if err = util.CopyFile(groupName, path.Join(fullPath, groupName)); err != nil {
		return err
	}
	if err = appendLines(passwdName, passwdEntries); err != nil {
		return err
	}
	if err = appendLines(groupName, groupEntries); err != nil {
		return err
	}
	return nil

}

/*
creates simple user db []simpleUserInfo  for a /etc/{passwd|group} file
*/
func createPasswdMap(fileName string) ([]simpleUserInfo, error) {
	var nameDb []simpleUserInfo
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fileScanner := bufio.NewScanner(file)
	for fileScanner.Scan() {
		line := fileScanner.Text()
		entries := strings.Split(line, ":")
		name := entries[0]
		uid, err := strconv.Atoi(entries[2])
		if err != nil {
			wwlog.Printf(wwlog.WARN, "could not parse uid(%s) for %s\n", entries[2], name)
		}
		gid, err := strconv.Atoi(entries[3])
		if err != nil {
			wwlog.Printf(wwlog.WARN, "could not parse gid(%s) for %s\n", entries[2], name)
		}
		if name != "" {
			nameDb = append(nameDb, simpleUserInfo{name: name, uid: uid, gid: gid})

		}
	}
	wwlog.Printf(wwlog.DEBUG, fmt.Sprintf("created uid/gid map with %v entries from %s\n", len(nameDb), fileName))
	return nameDb, nil
}

/*
Creates a slice with the entries of of passwd for the given slice of user names
*/
func getEntires(fileName string, names []string) ([]string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var entries []string
	fileScanner := bufio.NewScanner(file)
	for fileScanner.Scan() {
		line := fileScanner.Text()
		entries := strings.Split(line, ":")
		for _, name := range names {
			if entries[0] == name {
				entries = append(entries, line)
			}
		}
	}
	return entries, nil

}

/*
Appending the lines to the given file
*/
func appendLines(fileName string, lines []string) error {
	wwlog.Printf(wwlog.VERBOSE, "appending %v lines to %s\n", len(lines), fileName)
	file, err := os.OpenFile(fileName, os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, line := range lines {
		if _, err := file.WriteString(line); err != nil {
			return err
		}

	}
	return nil
}
