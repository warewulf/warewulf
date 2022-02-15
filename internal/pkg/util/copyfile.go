package util

import (
	"io"
	"os"

	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

func CopyFile(src string, dst string) error {

	wwlog.Printf(wwlog.DEBUG, "Copying '%s' to '%s'\n", src, dst)

	// Open source file
	srcFD, err := os.Open(src)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open source file %s: %s\n", src, err)
		return err
	}
	defer srcFD.Close()

	// Confirm source file structure is readable
	_, err = srcFD.Stat()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not stat source file %s: %s\n", src, err)
		return err
	}

	dstFD, err := os.Create(dst)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not create destination file %s: %s\n", dst, err)
		return err
	}
	defer dstFD.Close()

	bytes, err := io.Copy(srcFD, dstFD)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "File copy from %s to %s failed.\n %s\n", src, dst, err)
		return err
	} else {
		wwlog.Printf(wwlog.DEBUG, "Copied %d bytes from %s to %s.\n", bytes, src, dst)
	}

	err = CopyUIDGID(src, dst)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Ownership copy from %s to %s failed.\n %s\n", src, dst, err)
		return err
	}
	return nil
}
