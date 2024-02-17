package warewulfd

import (
	"fmt"
	"os"
	"path"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"

	"github.com/warewulf/warewulf/internal/pkg/container"
	"github.com/warewulf/warewulf/internal/pkg/util"
)

/*
Copies the default shim, which is the shim located on host
to the tftp directory
*/

func CopyShimGrub() (err error) {
	conf := warewulfconf.Get()
	wwlog.Debug("copy shim and grub binaries from host")
	shimPath := container.ShimFind("")
	if shimPath == "" {
		return fmt.Errorf("no shim found on the host os")
	}
	err = util.CopyFile(shimPath, path.Join(conf.TFTP.TftpRoot, "warewulf", "shim.efi"))
	if err != nil {
		return err
	}
	_ = os.Chmod(path.Join(conf.TFTP.TftpRoot, "warewulf", "shim.efi"), 0o755)
	grubPath := container.GrubFind("")
	if grubPath == "" {
		return fmt.Errorf("no grub found on host os")
	}
	err = util.CopyFile(grubPath, path.Join(conf.TFTP.TftpRoot, "warewulf", "grub.efi"))
	if err != nil {
		return err
	}
	_ = os.Chmod(path.Join(conf.TFTP.TftpRoot, "warewulf", "grub.efi"), 0o755)
	err = util.CopyFile(grubPath, path.Join(conf.TFTP.TftpRoot, "warewulf", "grubx64.efi"))
	_ = os.Chmod(path.Join(conf.TFTP.TftpRoot, "warewulf", "grubx64.efi"), 0o755)

	return
}
