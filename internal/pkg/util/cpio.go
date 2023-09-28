package util

import (
	"io"
	"os"

	"github.com/sassoftware/go-rpmutils/cpio"
)

/*
Opens cpio archive and returns the file list
*/
func CpioFiles(name string) (files []string, err error) {
	f, err := os.Open(name)
	if err != nil {
		return files, err
	}
	defer f.Close()

	reader := cpio.NewReader(f)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return files, nil
		}
		if err != nil {
			return files, err
		}
		files = append(files, header.Filename())
	}
}
