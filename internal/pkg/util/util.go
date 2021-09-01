package util

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"syscall"
	"time"

	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

func DirModTime(path string) (time.Time, error) {

	var lastTime time.Time
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		cur := info.ModTime()
		if cur.After(lastTime) {
			lastTime = info.ModTime()
		}

		return nil
	})
	if err != nil {
		return time.Time{}, err
	}
	return lastTime, nil
}

func PathIsNewer(source string, compare string) bool {
	wwlog.Printf(wwlog.DEBUG, "Comparing times on paths: '%s' - '%s'\n", source, compare)

	time1, err := DirModTime(source)
	if err != nil {
		wwlog.Printf(wwlog.DEBUG, "%s\n", err)
		return false
	}

	time2, err := DirModTime(compare)
	if err != nil {
		wwlog.Printf(wwlog.DEBUG, "%s\n", err)
		return false
	}

	return time1.Before(time2)
}

func RandomString(n int) string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

func CopyFile(source string, dest string) error {
	wwlog.Printf(wwlog.DEBUG, "Copying '%s' to '%s'\n", source, dest)
	sourceFD, err := os.Open(source)
	if err != nil {
		return err
	}

	finfo, err := sourceFD.Stat()

	destFD, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, finfo.Mode())
	if err != nil {
		return err
	}

	_, err = io.Copy(destFD, sourceFD)
	if err != nil {
		return err
	}

	CopyUIDGID(source, dest)
	if err != nil {
		return err
	}
	sourceFD.Close()

	return destFD.Close()
}

func CopyFiles(source string, dest string) error {
	err := filepath.Walk(source, func(location string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			wwlog.Printf(wwlog.DEBUG, "Creating directory: %s\n", location)
			info, err := os.Stat(source)
			if err != nil {
				return err
			}

			err = os.MkdirAll(path.Join(dest, location), info.Mode())
			if err != nil {
				return err
			}
			err = CopyUIDGID(source,dest)
			if err != nil {
				return err
			}

		} else {
			wwlog.Printf(wwlog.DEBUG, "Writing file: %s\n", location)

			err := CopyFile(location, path.Join(dest, location))
			if err != nil {
				return err
			}

		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

//TODO: func CopyRecursive ...

func IsDir(path string) bool {
	wwlog.Printf(wwlog.DEBUG, "Checking if path exists as a directory: %s\n", path)

	if path == "" {
		return false
	}
	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		return true
	}
	return false
}

func IsFile(path string) bool {
	wwlog.Printf(wwlog.DEBUG, "Checking if path exists as a file: %s\n", path)

	if path == "" {
		return false
	}

	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func ReadFile(path string) ([]string, error) {
	lines := []string{}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	f.Close()
	return lines, nil
}

func ValidString(pattern string, expr string) bool {
	if b, _ := regexp.MatchString(expr, pattern); b == true {
		return true
	}
	return false
}

func ValidateOrDie(message string, pattern string, expr string) {
	if ValidString(pattern, expr) == false {
		wwlog.Printf(wwlog.ERROR, "%s does not validate: '%s'\n", message, pattern)
		os.Exit(1)
	}
}

func FindFiles(path string) []string {
	var ret []string

	wwlog.Printf(wwlog.DEBUG, "Changing directory to FindFiles path: %s\n", path)
	err := os.Chdir(path)
	if err != nil {
		wwlog.Printf(wwlog.WARN, "Could not chdir() to: %s\n", path)
		return ret
	}

	err = filepath.Walk(".", func(location string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if location == "." {
			return nil
		}

		if IsDir(location) == true {
			wwlog.Printf(wwlog.DEBUG, "FindFiles() found directory: %s\n", location)
			ret = append(ret, location+"/")
		} else {
			wwlog.Printf(wwlog.DEBUG, "FindFiles() found file: %s\n", location)
			ret = append(ret, location)
		}

		return nil
	})
	if err != nil {
		return ret
	}

	return ret
}

func ExecInteractive(command string, a ...string) error {
	wwlog.Printf(wwlog.DEBUG, "ExecInteractive(%s, %s)\n", command, a)
	c := exec.Command(command, a...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err := c.Run()
	return err
}

func ShaSumFile(file string) (string, error) {
	var ret string

	f, err := os.Open(file)
	if err != nil {
		return ret, nil
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return ret, err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func SliceRemoveElement(array []string, remove string) []string {
	var ret []string

	// Linear time, maintains order
	for _, r := range array {
		if r != remove {
			ret = append(ret, r)
		} else {
			wwlog.Printf(wwlog.DEBUG, "Removing slice from array: %s\n", remove)
		}
	}

	return ret
}

func SliceAddUniqueElement(array []string, add string) []string {
	var ret []string
	var found bool

	//Linear time, appends
	for _, r := range array {
		ret = append(ret, r)
		if r == add {
			found = true
		}
	}

	if found == false {
		ret = append(ret, add)
	}

	return ret
}

func SystemdStart(systemdName string) error {
	startCmd := fmt.Sprintf("systemctl restart %s", systemdName)
	enableCmd := fmt.Sprintf("systemctl enable %s", systemdName)

	wwlog.Printf(wwlog.DEBUG, "Setting up Systemd service: %s\n", systemdName)
	ExecInteractive("/bin/sh", "-c", startCmd)
	ExecInteractive("/bin/sh", "-c", enableCmd)

	return nil
}

func CopyUIDGID(source string, dest string) error {
	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	// root is always good, if we failt to get UID/GID of a file
	var UID int = 0
	var GID int = 0
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		UID = int(stat.Uid)
		GID = int(stat.Gid)
	}
	wwlog.Printf(wwlog.DEBUG, "Chown '%i':'%i' '%s'\n", UID, GID, dest)
	err = os.Chown(dest, UID, GID)
	return err
}
