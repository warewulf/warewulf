package kernel

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

var kernelBuildTests = []struct {
	kernelVersion  string
	kernelName     string
	kernelFileName string
	succeed        bool
}{
	// kernel naming convention is <base kernel version>-<ABI number>.<upload number>-<flavour>
	{"4.3.2-1", "kernel1", "vmlinuz-1.2.3-4.gz", false},
	{"1.2.3-4", "kernel1", "vmlinuz-1.2.3-4.gz", true},
	{"1.2.3-4.3.1-generic", "kernel1", "vmlinuz-1.2.3-4.3.1-generic.gz", true},
}

func Test_BuildKernel(t *testing.T) {
	wwlog.SetLogLevel(wwlog.DEBUG)
	for _, tt := range kernelBuildTests {
		srvDir, err := os.MkdirTemp(os.TempDir(), "ww-test-srv-*")
		assert.NoError(t, err)
		conf := warewulfconf.Get()
		conf.Paths.WWProvisiondir = srvDir
		kernelDir, err := os.MkdirTemp(os.TempDir(), "ww-test-kernel-*")
		assert.NoError(t, err)
		{
			err = os.MkdirAll(path.Join(kernelDir, "boot"), 0755)
			assert.NoError(t, err)
			err = os.MkdirAll(path.Join(kernelDir, "lib/modules/old-kernel"), 0755)
			assert.NoError(t, err)
			_, err = os.Create(path.Join(kernelDir, "lib/modules/old-kernel/old-module"))
			assert.NoError(t, err)
			err = os.MkdirAll(path.Join(kernelDir, "lib/firmware"), 0755)
			assert.NoError(t, err)
			_, err = os.Create(path.Join(kernelDir, "lib/firmware/test-firmware"))
			assert.NoError(t, err)
			_, err = os.Create(path.Join(kernelDir, "boot", tt.kernelFileName))
			assert.NoError(t, err)
			err = os.MkdirAll(path.Join(kernelDir, "lib/modules", tt.kernelVersion, "/nested"), 0755)
			assert.NoError(t, err)
			_, err = os.Create(path.Join(kernelDir, "lib/modules", tt.kernelVersion, "test-module"))
			assert.NoError(t, err)
			err = os.Symlink(path.Join(kernelDir, "lib/modules/old-kernel/old-module"), path.Join(kernelDir, "lib/modules", tt.kernelVersion, "symlink-module"))
			assert.NoError(t, err)
		}
		t.Run(tt.kernelName, func(t *testing.T) {
			err = Build(tt.kernelVersion, tt.kernelName, kernelDir)
			if tt.succeed {
				assert.NoError(t, err)
				assert.FileExists(t, path.Join(srvDir, "kernel", tt.kernelName, "vmlinuz"))
				assert.FileExists(t, path.Join(srvDir, "kernel", tt.kernelName, "kmods.img.gz"))
				assert.FileExists(t, path.Join(srvDir, "kernel", tt.kernelName, "kmods.img"))
				files, err := util.CpioFiles(path.Join(srvDir, "kernel", tt.kernelName, "kmods.img"))
				assert.NoError(t, err)
				assert.ElementsMatch(t, files, []string{
					"lib/firmware/test-firmware",
					"lib/modules/" + tt.kernelVersion + "/symlink-module",
					"lib/modules/" + tt.kernelVersion + "/test-module",
					"lib/modules/" + tt.kernelVersion + "/nested"})
			} else {
				assert.Error(t, err)
			}
		})
		os.RemoveAll(srvDir)
		os.RemoveAll(kernelDir)
	}
}

var kernelFindTests = []struct {
	name        string
	prefix      string // kernel name prefix
	kernelNames []string
	expVer      string
	expPath     string
}{
	{
		name:        "vmlinuz under boot directory ok case",
		prefix:      "/boot/vmlinuz-%s",
		kernelNames: []string{"5.14.0-427.18.1.el9_4.x86_64", "5.14.0-427.24.1.el9_4.x86_64", "4.14.0-427.18.1.el8_4.x86_64"},
		expVer:      "5.14.0-427.24.1.el9_4.x86_64",
		expPath:     "/boot/vmlinuz-5.14.0-427.24.1.el9_4.x86_64",
	},
	{
		name:        "vmlinuz under boot directory ok case 2",
		prefix:      "/boot/vmlinuz-%s",
		kernelNames: []string{"5.15.0-119-generic", "5.14.0-427.24.1.el9_4.x86_64", "6.15.0-119-generic"},
		expVer:      "6.15.0-119-generic",
		expPath:     "/boot/vmlinuz-6.15.0-119-generic",
	},
	{
		name:        "vmlinuz under boot directory ok case 3",
		prefix:      "/boot/vmlinuz-%s",
		kernelNames: []string{"5.15.0-0-vanilla", "5.14.0-427.24.1.el9_4.x86_64"},
		expVer:      "5.15.0-0-vanilla",
		expPath:     "/boot/vmlinuz-5.15.0-0-vanilla",
	},
	{
		// <base kernel version>-<ABI number>.<upload number>-<flavour>
		name:        "vmlinuz under boot directory ok case (becuase the first version naming is incorrect)",
		prefix:      "/boot/vmlinuz-%s",
		kernelNames: []string{"5.15.0-generic", "5.14.0-427.24.1.el9_4.x86_64", "5.13.0-427.24.1.el9_4.x86_64"},
		expVer:      "5.14.0-427.24.1.el9_4.x86_64",
		expPath:     "/boot/vmlinuz-5.14.0-427.24.1.el9_4.x86_64",
	},
	{
		name:        "vmlinuz under lib modules ok case",
		prefix:      "/lib/modules/%s/vmlinuz",
		kernelNames: []string{"5.14.0-427.18.1.el9_4.x86_64", "5.14.0-427.24.1.el9_4.x86_64"},
		expVer:      "5.14.0-427.24.1.el9_4.x86_64",
		expPath:     "/lib/modules/5.14.0-427.24.1.el9_4.x86_64/vmlinuz",
	},
	{
		name:        "vmlinuz.gz under boot directory ok case",
		prefix:      "/boot/vmlinuz-%s.gz",
		kernelNames: []string{"5.14.0-427.18.1.el9_4.x86_64", "5.14.0-427.24.1.el9_4.x86_64"},
		expVer:      "5.14.0-427.24.1.el9_4.x86_64",
		expPath:     "/boot/vmlinuz-5.14.0-427.24.1.el9_4.x86_64.gz",
	},
	{
		name:        "mixed rescue / debug kernel testing",
		prefix:      "/boot/vmlinuz-%s",
		kernelNames: []string{"0-rescue-eb46964329b146e39518c625feab3ea0", "5.14.0-362.24.1.el9_3.aarch64", "5.14.0-427.31.1.el9_4.aarch64+debug", "5.14.0-284.30.1.el9_2.aarch64", "5.14.0-427.31.1.el9_4.aarch64"},
		expVer:      "5.14.0-427.31.1.el9_4.aarch64",
		expPath:     "/boot/vmlinuz-5.14.0-427.31.1.el9_4.aarch64",
	},
}

func Test_FindKernel(t *testing.T) {
	wwlog.SetLogLevel(wwlog.DEBUG)
	for _, tt := range kernelFindTests {
		srvDir, err := os.MkdirTemp(os.TempDir(), "ww-test-srv-*")
		assert.NoError(t, err)
		conf := warewulfconf.Get()
		conf.Paths.WWProvisiondir = srvDir
		kernelDir, err := os.MkdirTemp(os.TempDir(), "ww-test-kernel-*")
		assert.NoError(t, err)
		{
			for _, version := range tt.kernelNames {
				kernel := fmt.Sprintf(tt.prefix, version)
				parent := filepath.Dir(kernel)
				err = os.MkdirAll(path.Join(kernelDir, parent), 0755)
				assert.NoError(t, err)
				_, err := os.Create(path.Join(kernelDir, kernel))
				assert.NoError(t, err)
			}
		}

		t.Run(tt.name, func(t *testing.T) {
			kpath, kver, err := FindKernel(kernelDir)
			assert.NoError(t, err)
			assert.Equal(t, tt.expVer, kver)
			assert.Equal(t, filepath.Join(kernelDir, tt.expPath), kpath)
		})
		os.RemoveAll(srvDir)
		os.RemoveAll(kernelDir)
	}
}
