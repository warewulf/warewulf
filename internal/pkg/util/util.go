package util

import (
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"time"

	//   "strings"
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

	sourceFD.Close()

	return destFD.Close()
}


func IsDir(path string) (bool) {
	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		return true
	}
	return false
}


func TaintCheck(pattern string, expr string) bool {
	if b, _ := regexp.MatchString(expr, pattern); b == true {
		return true
	}
	return false
}

func ValidateOrDie(hostname string, name string, pattern string, expr string) {
	if TaintCheck(pattern, expr) == false {
		wwlog.Printf(wwlog.ERROR, "Entry '%s:%s' contains illegal characters: '%s'\n", hostname, name, pattern)
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

		if IsDir(location) == false {
			wwlog.Printf(wwlog.DEBUG, "FindFiles() found: %s\n", location)
			ret = append(ret, location)
		}
		return nil
	})
	if err != nil {
		return ret
	}

	return ret
}