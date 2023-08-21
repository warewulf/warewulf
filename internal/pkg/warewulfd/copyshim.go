package warewulfd

import (
	"fmt"
	"os"
	"path"

	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"

	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
)

/*
Copies the default shim, which is the shim located in the default container
to the tftp directory
*/

func CopyShimGrub() (err error) {
	wwlog.Debug("copy shim and grub binaries")
	nodeDB, err := node.New()
	if err != nil {
		return err
	}
	conf := warewulfconf.Get()
	profiles, err := nodeDB.MapAllProfiles()
	if err != nil {
		return err
	}
	if _, ok := profiles["default"]; !ok {
		return fmt.Errorf("default profile doesn't exist")
	}
	// *Entry.Get doesn't work the same as it works for nodes!
	if profiles["default"].BootMethod.Get() == "ipxe" || profiles["default"].BootMethod.Get() == "" {
		wwlog.Verbose("default profile uses ipxe boot")
		return
	}
	shimPath := container.ShimFind(profiles["default"].ContainerName.Get())
	if shimPath == "" {
		return fmt.Errorf("no shim found in the container: %s", profiles["default"].ContainerName.Get())
	}
	err = util.CopyFile(shimPath, path.Join(conf.Paths.Tftpdir, "warewulf", "shim.efi"))
	if err != nil {
		return err
	}
	_ = os.Chmod(path.Join(conf.Paths.Tftpdir, "warewulf", "shim.efi"), 0o755)
	grubPath := container.GrubFind(profiles["default"].ContainerName.Get())
	if grubPath == "" {
		return fmt.Errorf("no grub found in the container: %s", profiles["default"].ContainerName.Get())
	}
	err = util.CopyFile(grubPath, path.Join(conf.Paths.Tftpdir, "warewulf", "grub.efi"))
	if err != nil {
		return err
	}
	_ = os.Chmod(path.Join(conf.Paths.Tftpdir, "warewulf", "grub.efi"), 0o755)
	err = util.CopyFile(grubPath, path.Join(conf.Paths.Tftpdir, "warewulf", "grubx64.efi"))
	_ = os.Chmod(path.Join(conf.Paths.Tftpdir, "warewulf", "grubx64.efi"), 0o755)

	return
}
