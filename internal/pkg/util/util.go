package util

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"syscall"
	"time"

	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
)

func DirModTime(path string) (time.Time, error) {

	var lastTime time.Time
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		fi, err := os.Stat(path)
		if err != nil {
			return nil
		}
		stat := fi.Sys().(*syscall.Stat_t)
		cur := time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))
		if cur.After(lastTime) {
			lastTime = time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))
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

/*
Checks if given string is in slice. I yes returns true, false otherwise.
*/
func InSlice(slice []string, match string) bool {
	for _, val := range slice {
		if val == match {
			return true
		}
	}
	return false
}

/*
Checks if one or more elements of a slice A are a part of slice B. Returns true
as soon as one element matches.\
*/
func SliceInSlice(A []string, B []string) bool {
	for _, a := range A {
		for _, b := range B {
			if a == b {
				return true
			}
		}
	}
	return false
}

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

	if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
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
	if b, _ := regexp.MatchString(expr, pattern); b {
		return true
	}
	return false
}

func ValidateOrDie(message string, pattern string, expr string) {
	if ValidString(pattern, expr) {
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

		if IsDir(location) {
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

/*
Adds a string, to string slice if the given string is not present in the slice.
*/
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

	if !found {
		ret = append(ret, add)
	}

	return ret
}

/*
Appends a string slice to another slice. Guarantess that the elements are uniq.
*/
func SliceAppendUniq(array []string, add []string) []string {
	ret := array
	for _, r := range add {
		ret = SliceAddUniqueElement(ret, r)
	}
	return ret
}

func SystemdStart(systemdName string) error {
	startCmd := fmt.Sprintf("systemctl restart %s", systemdName)
	enableCmd := fmt.Sprintf("systemctl enable %s", systemdName)

	wwlog.Printf(wwlog.DEBUG, "Setting up Systemd service: %s\n", systemdName)
	err := ExecInteractive("/bin/sh", "-c", startCmd)
	if err != nil {
		return errors.Wrap(err, "failed to run start cmd")
	}
	err = ExecInteractive("/bin/sh", "-c", enableCmd)
	if err != nil {
		return errors.Wrap(err, "failed to run enable cmd")
	}

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
	wwlog.Printf(wwlog.DEBUG, "Chown %d:%d '%s'\n", UID, GID, dest)
	err = os.Chown(dest, UID, GID)
	return err
}

func SplitEscaped(input, delim, escape string) []string {
	var ret []string
	str := ""
	for i := 1; i < len(input); i++ {
		str += string(input[i-1])
		if string(input[i]) == delim && string(input[i-1]) != escape {
			i++
			ret = append(ret, str)
			str = ""
		}
		if string(input[i]) == escape {
			i++
		}

	}
	str += string(input[len(input)-1])
	ret = append(ret, str)

	return (ret)
}

func SplitValidPaths(input, delim string) []string {
	var ret []string
	str := ""
	for i := 1; i < len(input); i++ {
		str += string(input[i-1])
		if (string(input[i]) == delim && string(input[i-1]) != "\\") && (IsDir(str) || IsFile(str)) {
			i++
			ret = append(ret, str)
			str = ""
		}
		if string(input[i]) == "\\" {
			i++
		}

	}
	str += string(input[len(input)-1])
	ret = append(ret, str)

	return (ret)
}

func IncrementIPv4(start string, inc uint) string {
	ip_start := net.ParseIP(start)
	ipv4 := ip_start.To4()
	v4_int := uint(ipv4[0])<<24 + uint(ipv4[1])<<16 + uint(ipv4[2])<<8 + uint(ipv4[3])
	v4_int += inc
	v4_o3 := byte(v4_int & 0xFF)
	v4_o2 := byte((v4_int >> 8) & 0xFF)
	v4_o1 := byte((v4_int >> 16) & 0xFF)
	v4_o0 := byte((v4_int >> 24) & 0xFF)
	ipv4_new := net.IPv4(v4_o0, v4_o1, v4_o2, v4_o3)
	return ipv4_new.String()
}

/*
Appending the lines to the given file
*/
func AppendLines(fileName string, lines []string) error {
	wwlog.Printf(wwlog.VERBOSE, "appending %v lines to %s\n", len(lines), fileName)
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Errorf("Can't open file %s: %s", fileName, err)
	}
	defer file.Close()
	for _, line := range lines {
		wwlog.Printf(wwlog.DEBUG, "Appending '%s' to %s\n", line, fileName)
		if _, err := file.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
			return errors.Errorf("Can't write to file %s: %s", fileName, err)
		}

	}
	return nil
}
