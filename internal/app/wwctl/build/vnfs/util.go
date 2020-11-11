package vnfs


import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/vnfs"
	"os"
	"os/exec"
	"path"
)

func buildVnfs(source string, dest string) error {
	cmd := fmt.Sprintf("cd %s; find . | cpio --quiet -o -H newc | gzip -c > \"%s\"", source, dest)

	err := exec.Command("/bin/sh", "-c", cmd).Run()

	return err
}

func buildLinks(v vnfs.VnfsObject, source string) error {

	// Just incase the temporary link location is present, remove it if we can
	_ = os.Remove(v.Root() + "-link")

	// Just incase the directory doesn't exist, make it
	_ = os.MkdirAll(path.Dir(v.Root()), 0755)

	// Link to a temporary location so we can atomically move the link into place
	err := os.Symlink(source, v.Root()+"-link")
	if err != nil {
		return err
	}

	// Atomically move the link into place
	err = os.Rename(v.Root()+"-link", v.Root())
	if err != nil {
		return err
	}

	return nil
}
