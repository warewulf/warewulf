package container

import (
	"os"
	"path"
	"testing"

	// warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"

	"github.com/stretchr/testify/assert"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_Find_ShimX86(t *testing.T) {
	testenv.New(t)
	conf := warewulfconf.Get()
	wwlog.SetLogLevel(wwlog.DEBUG)
	os.MkdirAll(path.Join(conf.Paths.WWChrootdir, "suse/rootfs/usr/lib64/efi/"), 0755)
	shimF, err := os.Create(path.Join(conf.Paths.WWChrootdir, "suse/rootfs/usr/lib64/efi/shim.efi"))
	assert.NoError(t, err)
	shimF.WriteString("shim.efi")
	assert.FileExists(t, path.Join(conf.Paths.WWChrootdir, "suse/rootfs/usr/lib64/efi/shim.efi"))
	shimPath := ShimFind("suse")
	assert.Equal(t, path.Join(conf.Paths.WWChrootdir, "suse/rootfs/usr/lib64/efi/shim.efi"), shimPath)
}
func Test_Find_ShimArch64(t *testing.T) {
	testenv.New(t)
	conf := warewulfconf.Get()
	wwlog.SetLogLevel(wwlog.DEBUG)
	os.MkdirAll(path.Join(conf.Paths.WWChrootdir, "suse/rootfs/usr/share/efi/aarch64"), 0755)
	shimF, err := os.Create(path.Join(conf.Paths.WWChrootdir, "suse/rootfs/usr/share/efi/aarch64/shim.efi"))
	assert.NoError(t, err)
	shimF.WriteString("shim.efi")
	assert.FileExists(t, path.Join(conf.Paths.WWChrootdir, "suse/rootfs/usr/share/efi/aarch64/shim.efi"))
	shimPath := ShimFind("suse")
	assert.Equal(t, path.Join(conf.Paths.WWChrootdir, "suse/rootfs/usr/share/efi/aarch64/shim.efi"), shimPath)
}
func Test_Find_GrubX86(t *testing.T) {
	testenv.New(t)
	conf := warewulfconf.Get()
	wwlog.SetLogLevel(wwlog.DEBUG)
	os.MkdirAll(path.Join(conf.Paths.WWChrootdir, "suse/rootfs/usr/share/efi/x86_64"), 0755)
	shimF, err := os.Create(path.Join(conf.Paths.WWChrootdir, "suse/rootfs/usr/share/efi/x86_64/grub.efi"))
	assert.NoError(t, err)
	shimF.WriteString("grub.efi")
	assert.FileExists(t, path.Join(conf.Paths.WWChrootdir, "suse/rootfs//usr/share/efi/x86_64/grub.efi"))
	shimPath := GrubFind("suse")
	assert.Equal(t, path.Join(conf.Paths.WWChrootdir, "suse/rootfs/usr/share/efi/x86_64/grub.efi"), shimPath)
}
func Test_Find_GrubAarch64(t *testing.T) {
	testenv.New(t)
	conf := warewulfconf.Get()
	wwlog.SetLogLevel(wwlog.DEBUG)
	os.MkdirAll(path.Join(conf.Paths.WWChrootdir, "suse/rootfs/usr/share/efi/aarch64/"), 0755)
	shimF, err := os.Create(path.Join(conf.Paths.WWChrootdir, "suse/rootfs/usr/share/efi/aarch64/grub.efi"))
	assert.NoError(t, err)
	shimF.WriteString("grub.efi")
	assert.FileExists(t, path.Join(conf.Paths.WWChrootdir, "suse/usr/share/efi/aarch64/grub.efi"))
	shimPath := GrubFind("suse")
	assert.Equal(t, path.Join(conf.Paths.WWChrootdir, "suse/rootfs/usr/share/efi/aarch64//grub.efi"), shimPath)
}
