package warewulfd

import (
	"fmt"
	"path"

	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"

	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
)

/*
Copies the default shim, which is the shim located in the default container
to the tftp directory
*/

func CopyShimGrub() (err error) {
	nodeDB, err := node.New()
	if err != nil {
		return err
	}
	conf := warewulfconf.Get()
	if err != nil {
		return err
	}
	profiles, err := nodeDB.MapAllProfiles()
	if err != nil {
		return err
	}
	if _, ok := profiles["default"]; ok {
		return fmt.Errorf("default profile doesn't exist")
	}
	if profiles["default"].BootMethod.Get() == "ipxe" {
		return
	}
	shimPath := container.ShimFind("default")
	if shimPath == "" {
		return fmt.Errorf("no shim found in the default profile")
	}
	err = util.CopyFile(shimPath, path.Join(conf.TFTP.TftpRoot, "shim.efi"))
	if err != nil {
		return err
	}
	grubPath := container.GrubFind("default")
	if shimPath == "" {
		return fmt.Errorf("no grub found in the default profile")
	}
	err = util.CopyFile(grubPath, path.Join(conf.TFTP.TftpRoot, "grub.efi"))
	return
}
