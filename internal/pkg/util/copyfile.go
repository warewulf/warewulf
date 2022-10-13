package util

import (
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

func CopyFile(src string, dst string) error {

	wwlog.Debug("Copying '%s' to '%s'", src, dst)

	// Open source file
	srcFD, err := os.Open(src)
	if err != nil {
		wwlog.Error("Could not open source file %s: %s", src, err)
		return err
	}
	defer srcFD.Close()

	srcInfo, err := srcFD.Stat()
	if err != nil {
		wwlog.Error("Could not stat source file %s: %s", src, err)
		return err
	}

	dstFD, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		wwlog.Error("Could not create destination file %s: %s", dst, err)
		return err
	}
	defer dstFD.Close()

	bytes, err := io.Copy(dstFD, srcFD)
	if err != nil {
		wwlog.Error("File copy from %s to %s failed.\n %s", src, dst, err)
		return err
	} else {
		wwlog.Debug("Copied %d bytes from %s to %s.", bytes, src, dst)
	}

	err = CopyUIDGID(src, dst)
	if err != nil {
		wwlog.Error("Ownership copy from %s to %s failed.\n %s", src, dst, err)
		return err
	}
	return nil
}

func SafeCopyFile(src string, dst string) error {
	var err error
	// Don't overwrite existing files -- should add force overwrite switch
	if _, err = os.Stat(dst); err == nil {
		wwlog.Debug("Destination file %s exists.", dst)
	} else {
		err = CopyFile(src, dst)
	}
	return err
}

func CopyFiles(source string, dest string) error {
	err := filepath.Walk(source, func(location string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			wwlog.Debug("Creating directory: %s", location)
			info, err := os.Stat(source)
			if err != nil {
				return err
			}

			err = os.MkdirAll(path.Join(dest, location), info.Mode())
			if err != nil {
				return err
			}
			err = CopyUIDGID(source, dest)
			if err != nil {
				return err
			}

		} else {
			wwlog.Debug("Writing file: %s", location)

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
