package util

import (
	"io"
	"math/rand"
	"os"
	"path/filepath"
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

func RandomString(n int) string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

func CopyFile(source string, dest string) error {
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
