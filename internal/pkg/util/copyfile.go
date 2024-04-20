package util

import (
	"io"
	"os"

	"github.com/warewulf/warewulf/internal/pkg/wwlog"
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
