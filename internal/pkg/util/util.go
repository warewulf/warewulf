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
	"path"
	"path/filepath"
	"regexp"
	"strings"
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
	wwlog.Debug("Comparing times on paths: '%s' - '%s'", source, compare)

	time1, err := DirModTime(source)
	if err != nil {
		wwlog.DebugExc(err, "")
		return false
	}

	time2, err := DirModTime(compare)
	if err != nil {
		wwlog.DebugExc(err, "")
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

func IsDir(path string) bool {
	wwlog.Debug("Checking if path exists as a directory: %s", path)

	if path == "" {
		return false
	}
	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		return true
	}
	return false
}

func IsFile(path string) bool {
	wwlog.Debug("Checking if path exists as a file: %s", path)

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
		wwlog.Error("%s does not validate: '%s'", message, pattern)
		os.Exit(1)
	}
}

//******************************************************************************
func FindFiles(path string) []string {
	var ret []string

	wwlog.Debug("Changing directory to FindFiles path: %s", path)
	err := os.Chdir(path)
	if err != nil {
		wwlog.Warn("Could not chdir() to: %s", path)
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
			wwlog.Debug("FindFiles() found directory: %s", location)
			ret = append(ret, location+"/")
		} else {
			wwlog.Debug("FindFiles() found file: %s", location)
			ret = append(ret, location)
		}

		return nil
	})
	if err != nil {
		return ret
	}

	return ret
}

//******************************************************************************
func FindFilterFiles(
	path string,
	include []string,
	ignore []string,
	ignore_xdev bool) ([]string, error) {

	wwlog.Debug("Finding files: %s", path)

	ofiles := []string{}

	cwd, err := os.Getwd()
	if err != nil {
		return ofiles, err
	}
	defer os.Chdir(cwd)

	err = os.Chdir(path)
	if err != nil {
		return ofiles, errors.Wrapf(err, "Failed to change path: %s", path)
	}

	files := []string{}

	for _, pattern := range include {

		_files, err := filepath.Glob(pattern)
		if err != nil {
			return ofiles, errors.Wrapf(err, "Failed to apply pattern: %s", pattern)
		}
		wwlog.Debug("Including pattern: %s -> %d matches", pattern, len(_files))

		files = append(files, _files...)
	}


	for i, pattern := range(ignore) {
		if strings.HasPrefix(pattern, "./") {
			ignore[i] = pattern[2:]
		}
		wwlog.Debug("Ignore pattern (%d): %s", i, ignore[i])
	}


	if ignore_xdev {
		wwlog.Debug("Ignoring cross-device (xdev) files")
	}

	path_stat, err := os.Stat(".")
	if err != nil {
		return ofiles, err
	}

	dev := path_stat.Sys().(*syscall.Stat_t).Dev

	for _, ifile := range files {
		if stat, err := os.Stat(ifile); err == nil && stat.IsDir() {
			// recursivly include from the matched directory

			num_init := len(ofiles)
			err = filepath.Walk(ifile, func(location string, info os.FileInfo, err error) error {
				var file string

				if err != nil {
					return err
				}

				if location == "." {
					return nil
				}

				if info.IsDir() {
					file = location + "/"
				} else {
					file = location
				}

				if ignore_xdev && info.Sys().(*syscall.Stat_t).Dev != dev {
					wwlog.Debug("Ignored (cross-device): %s", file)
					return nil
				}

				for i, pattern := range(ignore) {
					m, err := filepath.Match(pattern, location)
					if err != nil {
						return err
					}

					if m {
						wwlog.Debug("Ignored (%d): %s", i, file)
						return nil
					}
				}

				ofiles = append(ofiles, file)

				return nil
			})

			num_final := len(ofiles)
			wwlog.Debug("Included: %s -> %d files", ifile, num_final-num_init)

			if err != nil {
				return ofiles, err
			}
		}else{
			wwlog.Debug("Included: %s", ifile)
			ofiles = append(ofiles, ifile)
		}
	}

	return ofiles, err
}

//******************************************************************************
func ExecInteractive(command string, a ...string) error {
	wwlog.Debug("ExecInteractive(%s, %s)", command, a)
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
			wwlog.Debug("Removing slice from array: %s", remove)
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

	wwlog.Debug("Setting up Systemd service: %s", systemdName)
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
	wwlog.Debug("Chown %d:%d '%s'", UID, GID, dest)
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
	wwlog.Verbose("appending %v lines to %s", len(lines), fileName)
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrapf(err, "Can't open file: %s", fileName)
	}
	defer file.Close()
	for _, line := range lines {
		wwlog.Debug("Appending '%s' to %s", line, fileName)
		if _, err := file.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
			return errors.Wrapf(err, "Can't write to file: %s", fileName)
		}

	}
	return nil
}

/*******************************************************************************
	Create an archive using cpio
*/
func CpioCreate(
	ifiles []string,
	ofile string,
	format string,
	cpio_args ...string ) error {

	args := []string{
		"--quiet",
		"--create",
		"-H", format,
		"--file=" + ofile }

	args = append(args, cpio_args...)

	proc := exec.Command("cpio", args...)

	stdin, err := proc.StdinPipe()
	if err != nil {
		return err
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, strings.Join(ifiles, "\n"))
	}()

	out, err := proc.CombinedOutput()
	if len(out) > 0 {
		wwlog.Debug(string(out))
	}

	return err
}

/*******************************************************************************
	Compress a file using gzip or pigz
*/
func FileGz(
	file string ) error {

	file_gz := file + ".gz"

	if IsFile(file_gz) {
		err := os.Remove(file_gz)

		if err != nil {
			return errors.Wrapf(err, "Could not remove existing file: ", file_gz)
		}
	}

	compressor, err := exec.LookPath("pigz")
	if err != nil {
		wwlog.Verbose("Could not locate PIGZ")
		compressor = "gzip"
	}

	wwlog.Verbose("Using gz compressor: %s", compressor)

	proc := exec.Command(
		compressor,
		"--keep",
	 	file )

	out, err := proc.CombinedOutput()
	if len(out) > 0 {
		wwlog.Debug(string(out))
	}

	return err
}

/*******************************************************************************
	Create an archive using cpio
*/
func BuildFsImage(
	name string,
	rootfsPath string,
	imagePath string,
	include []string,
	ignore []string,
	ignore_xdev bool,
	format string,
	cpio_args ...string ) error {

	err := os.MkdirAll(path.Dir(imagePath), 0755)
	if err != nil {
		return errors.Wrapf(err, "Failed to create image directory for %s: %s", name, imagePath)
	}

	wwlog.Debug("Created image directory for %s: %s", name, imagePath)

	// TODO: why is this done if the container must already exist?
	err = os.MkdirAll(path.Dir(rootfsPath), 0755)
	if err != nil {
		return errors.Wrapf(err, "Failed to create fs directory for %s: %s", name, rootfsPath)
	}

	wwlog.Debug("Created fs directory for %s: %s", name, rootfsPath)

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(cwd)

	err = os.Chdir(rootfsPath)
	if err != nil {
		return errors.Wrapf(err, "Failed chdir to fs directory for %s: %s", name, rootfsPath)
	}

	files, err := FindFilterFiles(
		".",
		include,
		ignore,
		ignore_xdev )
	if err != nil {
		return errors.Wrapf(err, "Failed discovering files for %s: %s", name, rootfsPath)
	}

	err = CpioCreate(
		files,
	 	imagePath,
		format,
 		cpio_args...)
	if err != nil {
		return errors.Wrapf(err, "Failed creating image for %s: %s", name, imagePath)
	}

	wwlog.Info("Created image for %s: %s", name, imagePath)

	err = FileGz(imagePath)
	if err != nil {
		return errors.Wrapf(err, "Failed to compress image for %s: %s", name, imagePath + ".gz")
	}

	wwlog.Info("Compressed image for %s: %s", name, imagePath + ".gz")

	return nil
}
