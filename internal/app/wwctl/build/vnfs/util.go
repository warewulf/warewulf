package vnfs

import (
	"fmt"
	"os/exec"
)

func buildVnfs(source string, dest string) error {
	cmd := fmt.Sprintf("cd %s; find . | cpio --quiet -o -H newc | gzip -c > \"%s\"", source, dest)

	err := exec.Command("/bin/sh", "-c", cmd).Run()

	return err
}
